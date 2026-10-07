// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

// Package workbook reads and writes .xlsx workbooks without a spreadsheet
// application. It is the single implementation behind the xlsx actions, the
// dagu xlsx command, and the MCP workbook read target, so every consumer sees
// the same typing, header detection, and change summaries.
package workbook

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// DefaultMaxRows is the number of rows a read returns when max_rows is unset.
const DefaultMaxRows = 5000

// Row is one sheet row keyed by header. Values are nil, string, int64,
// float64, or bool; dates are ISO 8601 strings. RowNumberKey holds the
// 1-based sheet row as an int.
type Row map[string]any

// RowNumberKey is the row field carrying the 1-based sheet row number.
const RowNumberKey = "_row"

// CellError names the workbook, sheet, and cell of a problem.
type CellError struct {
	Workbook string
	Sheet    string
	Cell     string
	Msg      string
}

func (e *CellError) Error() string {
	if e.Cell == "" {
		return fmt.Sprintf("%s %s: %s", e.Workbook, e.Sheet, e.Msg)
	}
	return fmt.Sprintf("%s %s!%s: %s", e.Workbook, e.Sheet, e.Cell, e.Msg)
}

// SheetNotFoundError lists the sheets a workbook has when a name misses.
type SheetNotFoundError struct {
	Workbook string
	Sheet    string
	Present  []string
}

func (e *SheetNotFoundError) Error() string {
	return fmt.Sprintf("%s: sheet %q not found; sheets present: %s", e.Workbook, e.Sheet, strings.Join(e.Present, ", "))
}

// LockedError reports a workbook another program holds open.
type LockedError struct {
	Path string
}

func (e *LockedError) Error() string {
	return fmt.Sprintf("%s is open in another program; close it and retry", filepath.Base(e.Path))
}

// NotFoundError reports a missing workbook. It unwraps to fs.ErrNotExist.
type NotFoundError struct {
	Path string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: workbook not found", filepath.Base(e.Path))
}

// Unwrap lets errors.Is(err, fs.ErrNotExist) hold.
func (*NotFoundError) Unwrap() error { return fs.ErrNotExist }

// ErrUnsupportedFormat is wrapped into errors for files that are not .xlsx.
var ErrUnsupportedFormat = errors.New("only .xlsx and .xlsm workbooks are supported; save as .xlsx")

// ErrNotWorkbook is wrapped into errors for files that cannot be parsed.
var ErrNotWorkbook = errors.New("not a valid .xlsx workbook")

// ErrPassword is wrapped into errors for a protected workbook opened
// without the right password.
var ErrPassword = errors.New("workbook password is missing or incorrect")

// PasswordError reports a protected workbook whose password was missing
// or wrong. It unwraps to ErrPassword.
type PasswordError struct {
	Path string
}

func (e *PasswordError) Error() string {
	return filepath.Base(e.Path) + ": " + ErrPassword.Error()
}

// Unwrap lets errors.Is(err, ErrPassword) hold.
func (*PasswordError) Unwrap() error { return ErrPassword }

// oleHeader is the Compound File Binary signature of an encrypted OOXML
// workbook. excelize reports a bad password as an unsupported format, so
// this header, together with the EncryptionInfo stream, is what
// distinguishes protection from a file that is not a workbook.
var oleHeader = []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}

// encryptionInfoName is the UTF-16LE stream name of an encrypted package.
var encryptionInfoName = []byte{
	'E', 0, 'n', 0, 'c', 0, 'r', 0, 'y', 0, 'p', 0, 't', 0, 'i', 0, 'o', 0, 'n', 0,
	'I', 0, 'n', 0, 'f', 0, 'o', 0,
}

// CheckExtension rejects paths whose extension is not .xlsx or .xlsm.
func CheckExtension(path string) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xlsx", ".xlsm":
		return nil
	default:
		return fmt.Errorf("%s: %w", filepath.Base(path), ErrUnsupportedFormat)
	}
}

// file is an open workbook with the caches reads need.
type file struct {
	f        *excelize.File
	path     string
	base     string
	date1904 bool
	sheets   []string
	kinds    map[int]cellKind
	grids    map[string][][]string
	// dated memoizes styles derived for dates written into plain cells.
	dated map[datedKey]int
}

func open(path, password string) (*file, error) {
	if err := CheckExtension(path); err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, &NotFoundError{Path: path}
		}
		return nil, classifyError(path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s: is a directory", filepath.Base(path))
	}
	opts := excelize.Options{Password: password}
	f, err := excelize.OpenFile(path, opts)
	if err != nil {
		if locked := classifyError(path, err); locked != nil && errors.As(locked, new(*LockedError)) {
			return nil, locked
		}
		if passwordProtected(path) {
			return nil, &PasswordError{Path: path}
		}
		return nil, fmt.Errorf("%s: %w: %v", filepath.Base(path), ErrNotWorkbook, err)
	}
	w := &file{f: f, path: path, base: filepath.Base(path), kinds: map[int]cellKind{}, grids: map[string][][]string{}}
	w.sheets = f.GetSheetList()
	if props, err := f.GetWorkbookProps(); err == nil && props.Date1904 != nil {
		w.date1904 = *props.Date1904
	}
	return w, nil
}

// passwordProtected reports whether path is an encrypted OOXML workbook.
// It is checked only after opening fails. Only the compound-file header,
// the FAT entries that locate the directory, and the directory sectors are
// read; the encrypted package is left on disk.
func passwordProtected(path string) bool {
	f, err := os.Open(path) //nolint:gosec // path is the workbook the caller asked to open
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return cfbHasEncryptionInfo(f, info.Size())
}

const (
	cfbHeaderBytes  = 512
	cfbDirEntrySize = 128
	cfbDIFATInHead  = 109
	// An encrypted workbook's directory is a few entries. The cap stops a
	// corrupt sector chain from walking into the encrypted package.
	cfbMaxDirSectors = 32
	cfbMaxDIFAT      = 64

	cfbEndOfChain = 0xFFFFFFFE
)

// cfbHasEncryptionInfo reports whether the compound file's directory names
// an EncryptionInfo stream.
func cfbHasEncryptionInfo(r io.ReaderAt, size int64) bool {
	if size < cfbHeaderBytes {
		return false
	}
	header := make([]byte, cfbHeaderBytes)
	if _, err := r.ReadAt(header, 0); err != nil {
		return false
	}
	if !bytes.Equal(header[:len(oleHeader)], oleHeader) {
		return false
	}
	major := binary.LittleEndian.Uint16(header[0x1A:])
	var sectorSize, headerSize int
	switch binary.LittleEndian.Uint16(header[0x1E:]) {
	case 9:
		sectorSize = 512
		headerSize = cfbHeaderBytes
		if major != 3 {
			return false
		}
	case 12:
		sectorSize = 4096
		headerSize = sectorSize
		if major != 4 {
			return false
		}
	default:
		return false
	}
	c := cfbReader{
		r:          r,
		size:       size,
		header:     header,
		headerSize: headerSize,
		sectorSize: sectorSize,
	}
	dir := binary.LittleEndian.Uint32(header[0x30:])
	seen := make(map[uint32]struct{}, cfbMaxDirSectors)
	for range cfbMaxDirSectors {
		if _, ok := seen[dir]; ok || !c.sectorInFile(dir) {
			return false
		}
		seen[dir] = struct{}{}
		sector, err := c.readSector(dir)
		if err != nil {
			return false
		}
		for off := 0; off+cfbDirEntrySize <= len(sector); off += cfbDirEntrySize {
			if encryptionInfoEntry(sector[off : off+cfbDirEntrySize]) {
				return true
			}
		}
		next, ok := c.fatNext(dir)
		if !ok || next == cfbEndOfChain {
			return false
		}
		dir = next
	}
	return false
}

// encryptionInfoEntry reports whether a directory entry names the
// EncryptionInfo stream. The name is UTF-16LE and the length includes the
// terminating null.
func encryptionInfoEntry(entry []byte) bool {
	if len(entry) < cfbDirEntrySize {
		return false
	}
	n := int(binary.LittleEndian.Uint16(entry[64:66]))
	if n != len(encryptionInfoName)+2 {
		return false
	}
	return bytes.Equal(entry[:len(encryptionInfoName)], encryptionInfoName)
}

// cfbReader locates compound-file sectors without reading stream payloads.
type cfbReader struct {
	r          io.ReaderAt
	size       int64
	header     []byte
	headerSize int
	sectorSize int
}

func (c cfbReader) sectorInFile(sect uint32) bool {
	off, ok := c.sectorOffset(sect)
	return ok && off+int64(c.sectorSize) <= c.size
}

func (c cfbReader) sectorOffset(sect uint32) (int64, bool) {
	off := int64(c.headerSize) + int64(sect)*int64(c.sectorSize)
	if off < int64(c.headerSize) {
		return 0, false
	}
	return off, true
}

func (c cfbReader) readSector(sect uint32) ([]byte, error) {
	off, ok := c.sectorOffset(sect)
	if !ok {
		return nil, io.ErrUnexpectedEOF
	}
	buf := make([]byte, c.sectorSize)
	_, err := c.r.ReadAt(buf, off)
	return buf, err
}

// fatNext returns the next sector in a chain. The FAT itself is addressed
// through the header's DIFAT, one entry at a time.
func (c cfbReader) fatNext(sect uint32) (uint32, bool) {
	entries := uint32(c.sectorSize / 4)
	fatSect, ok := c.difat(sect / entries)
	if !ok || !c.sectorInFile(fatSect) {
		return 0, false
	}
	off, ok := c.sectorOffset(fatSect)
	if !ok {
		return 0, false
	}
	var buf [4]byte
	if _, err := c.r.ReadAt(buf[:], off+int64(sect%entries)*4); err != nil {
		return 0, false
	}
	return binary.LittleEndian.Uint32(buf[:]), true
}

// difat returns the sector number of the FAT sector that holds entry index.
func (c cfbReader) difat(index uint32) (uint32, bool) {
	if index < cfbDIFATInHead {
		at := 0x4C + int(index)*4
		return binary.LittleEndian.Uint32(c.header[at : at+4]), true
	}
	index -= cfbDIFATInHead
	sect := binary.LittleEndian.Uint32(c.header[0x44:])
	count := binary.LittleEndian.Uint32(c.header[0x48:])
	per := uint32(c.sectorSize/4 - 1)
	for n := uint32(0); n < count && n < cfbMaxDIFAT; n++ {
		if !c.sectorInFile(sect) {
			return 0, false
		}
		off, ok := c.sectorOffset(sect)
		if !ok {
			return 0, false
		}
		if index < per {
			var buf [4]byte
			if _, err := c.r.ReadAt(buf[:], off+int64(index)*4); err != nil {
				return 0, false
			}
			return binary.LittleEndian.Uint32(buf[:]), true
		}
		index -= per
		var next [4]byte
		if _, err := c.r.ReadAt(next[:], off+int64(per)*4); err != nil {
			return 0, false
		}
		sect = binary.LittleEndian.Uint32(next[:])
	}
	return 0, false
}

func (w *file) close() {
	if w != nil && w.f != nil {
		_ = w.f.Close()
	}
}

// resolveSheet returns the sheet a name refers to: the first sheet when the
// name is empty, an exact match, or a unique case-insensitive match.
func (w *file) resolveSheet(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		if len(w.sheets) == 0 {
			return "", fmt.Errorf("%s: workbook has no sheets", w.base)
		}
		return w.sheets[0], nil
	}
	for _, s := range w.sheets {
		if s == name {
			return s, nil
		}
	}
	match := ""
	for _, s := range w.sheets {
		if strings.EqualFold(s, name) {
			if match != "" {
				match = ""
				break
			}
			match = s
		}
	}
	if match != "" {
		return match, nil
	}
	return "", &SheetNotFoundError{Workbook: w.base, Sheet: name, Present: append([]string(nil), w.sheets...)}
}

func (w *file) cellError(sheet string, col, row int, msg string) *CellError {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	return &CellError{Workbook: w.base, Sheet: sheet, Cell: cell, Msg: msg}
}

func (w *file) sheetError(sheet, msg string) *CellError {
	return &CellError{Workbook: w.base, Sheet: sheet, Msg: msg}
}

// grid returns every cell of a sheet as raw strings indexed from row 1 and
// column 1 (index 0 is unused). Rows keep their own length, so a sparse
// sheet with one far-right cell does not allocate a full rectangle; cellAt
// treats a missing column as empty. Grids are cached per sheet for the life
// of the open workbook; writers must invalidate them.
func (w *file) grid(sheet string) ([][]string, error) {
	if cached, ok := w.grids[sheet]; ok {
		return cached, nil
	}
	rows, err := w.f.GetRows(sheet, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", w.base, sheet, err)
	}
	grid := make([][]string, len(rows)+1)
	grid[0] = nil
	for i, r := range rows {
		shifted := make([]string, len(r)+1)
		copy(shifted[1:], r)
		grid[i+1] = shifted
	}
	w.grids[sheet] = grid
	return grid, nil
}

// forget drops a sheet's cached grid after cells change.
func (w *file) forget(sheet string) {
	delete(w.grids, sheet)
}

func cellAt(grid [][]string, col, row int) string {
	if row < 1 || row >= len(grid) || col < 1 || col >= len(grid[row]) {
		return ""
	}
	return grid[row][col]
}

func rowIsEmpty(grid [][]string, row, c1, c2 int) bool {
	for c := c1; c <= c2; c++ {
		if strings.TrimSpace(cellAt(grid, c, row)) != "" {
			return false
		}
	}
	return true
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
