// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dagucloud/dagu/internal/cmd"
	"github.com/dagucloud/dagu/internal/core"
	"github.com/dagucloud/dagu/internal/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLsCommand(t *testing.T) {
	t.Run("ListsDAGsAndFiltersByPattern", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dagA := th.DAG(t, `name: ls-alpha
steps:
  - name: "1"
    run: echo a
`)
		_ = th.DAG(t, `name: ls-beta
steps:
  - name: "1"
    run: echo b
`)

		out := runLsWithStdout(t, th, cmd.Ls(), []string{"ls"})
		assert.Contains(t, out, "NAME")
		assert.Contains(t, out, "ls-alpha")
		assert.Contains(t, out, "ls-beta")

		out = runLsWithStdout(t, th, cmd.Ls(), []string{"ls", "alpha"})
		assert.Contains(t, out, dagA.Name)
		assert.NotContains(t, out, "ls-beta")
	})

	t.Run("ShowsLastAndHistoryColumns", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `name: ls-enriched
steps:
  - name: "1"
    run: echo hello
`)
		th.RunCommand(t, cmd.Start(), test.CmdTest{
			Args: []string{"start", dag.Location},
		})
		dag.AssertLatestStatus(t, core.Succeeded)

		out := runLsWithStdout(t, th, cmd.Ls(), []string{"ls", "-l", "-H", "ls-enriched"})
		assert.Contains(t, out, "LAST_STATUS")
		assert.Contains(t, out, "HISTORY")
		assert.Contains(t, out, "Succeeded")
	})

	t.Run("SortLastReverseHonorsNameTieBreaker", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		_ = th.DAG(t, `name: ls-tie-alpha
steps:
  - name: "1"
    run: echo a
`)
		_ = th.DAG(t, `name: ls-tie-beta
steps:
  - name: "1"
    run: echo b
`)

		// Neither DAG has run, so last-run times are equal (zero); reverse should
		// invert the case-insensitive name tie-breaker.
		outAsc := runLsWithStdout(t, th, cmd.Ls(), []string{"ls", "-t", "ls-tie-"})
		outDesc := runLsWithStdout(t, th, cmd.Ls(), []string{"ls", "-t", "-r", "ls-tie-"})

		assert.Less(t, strings.Index(outAsc, "ls-tie-alpha"), strings.Index(outAsc, "ls-tie-beta"))
		assert.Less(t, strings.Index(outDesc, "ls-tie-beta"), strings.Index(outDesc, "ls-tie-alpha"))
	})
}

// runLsWithStdout executes a cobra command and returns captured stdout.
func runLsWithStdout(t *testing.T, th test.Command, c *cobra.Command, args []string) string {
	t.Helper()

	var buf bytes.Buffer
	root := &cobra.Command{Use: "root"}
	root.AddCommand(c)
	root.SetOut(&buf)
	c.SetOut(&buf)
	root.SetArgs(test.WithConfigFlag(args, th.Config))

	err := root.ExecuteContext(th.Context)
	require.NoError(t, err)
	return buf.String()
}
