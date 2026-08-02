// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd

import (
	"testing"
	"time"

	"github.com/dagucloud/dagu/v2/internal/core"
	"github.com/stretchr/testify/assert"
)

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
