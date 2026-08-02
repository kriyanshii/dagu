// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/dagucloud/dagu/v2/internal/core"
	"github.com/dagucloud/dagu/v2/internal/core/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type warningDAGStore struct {
	exec.DAGStore
}

func (warningDAGStore) List(_ context.Context, opts exec.ListDAGsOptions) (exec.PaginatedResult[*core.DAG], []string, error) {
	return exec.NewPaginatedResult([]*core.DAG{}, 0, *opts.Paginator), []string{"catalog warning"}, nil
}

func TestRunLsWritesWarningsToCommandErrorStream(t *testing.T) {
	t.Parallel()

	command := Ls()
	command.SetOut(io.Discard)
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	err := runLs(&Context{
		Context:  context.Background(),
		Command:  command,
		DAGStore: warningDAGStore{},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, "warning: catalog warning\n", stderr.String())
}

func TestSortLsRowsByLastRun(t *testing.T) {
	t.Parallel()

	older := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	rows := []lsRow{
		{dag: &core.DAG{Name: "never-b"}},
		{dag: &core.DAG{Name: "older"}, lastTime: older},
		{dag: &core.DAG{Name: "newer"}, lastTime: newer},
		{dag: &core.DAG{Name: "never-a"}},
	}

	sortLsRowsByLastRun(rows, false)
	assert.Equal(t, []string{"newer", "older", "never-a", "never-b"}, lsRowNames(rows))

	sortLsRowsByLastRun(rows, true)
	assert.Equal(t, []string{"older", "newer", "never-b", "never-a"}, lsRowNames(rows))
}

func lsRowNames(rows []lsRow) []string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.dag.Name)
	}
	return names
}
