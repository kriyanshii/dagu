// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/dagucloud/dagu/internal/cmd"
	"github.com/dagucloud/dagu/internal/core"
	"github.com/dagucloud/dagu/internal/core/exec"
	"github.com/dagucloud/dagu/internal/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRmCommand(t *testing.T) {
	t.Run("RequiresHistoryOrDefinitionFlag", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)

		err := th.RunCommandWithError(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", dag.Name},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one of --history")
	})

	t.Run("DeletesAllHistoryWithForce", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)
		th.RunCommand(t, cmd.Start(), test.CmdTest{
			Args: []string{"start", dag.Location},
		})
		dag.AssertLatestStatus(t, core.Succeeded)
		dag.AssertDAGRunCount(t, 1)

		th.RunCommand(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "--history", "--force", dag.Name},
		})
		dag.AssertDAGRunCount(t, 0)
	})

	t.Run("OlderThanPreservesRecentHistory", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)
		th.RunCommand(t, cmd.Start(), test.CmdTest{
			Args: []string{"start", dag.Location},
		})
		dag.AssertLatestStatus(t, core.Succeeded)
		dag.AssertDAGRunCount(t, 1)

		th.RunCommand(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "-H", "-t", "30d", "-f", dag.Name},
		})
		dag.AssertDAGRunCount(t, 1)
	})

	t.Run("OlderThanRequiresHistory", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)

		err := th.RunCommandWithError(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "-d", "-t", "1d", "-f", dag.Name},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--older-than")
	})

	t.Run("DeletesDefinition", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)

		th.RunCommand(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "--definition", "--force", dag.Name},
		})

		_, err := th.DAGStore.GetMetadata(th.Context, dag.Location)
		require.Error(t, err)
	})

	t.Run("RefusesDefinitionDeleteWhenAlive", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		release := newHoldFile(t)
		dag := th.DAG(t, fmt.Sprintf(`steps:
  - name: "1"
    run: %q
`, holdUntilFileExistsCommand(release)))

		done := make(chan struct{})
		go func() {
			th.RunCommand(t, cmd.Start(), test.CmdTest{
				Args: []string{"start", dag.Location},
			})
			close(done)
		}()

		dag.AssertLatestStatus(t, core.Running)

		err := th.RunCommandWithError(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "-d", "-f", dag.Name},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "alive process")

		releaseHoldFile(t, release)
		<-done
	})

	t.Run("InvalidOlderThan", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `steps:
  - name: "1"
    run: echo "hello"
`)

		err := th.RunCommandWithError(t, cmd.Rm(), test.CmdTest{
			Args: []string{"rm", "-H", "-t", "bogus", "-f", dag.Name},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid --older-than")
	})
}

func TestPsCommand(t *testing.T) {
	t.Run("EmptyWhenNothingRunning", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		out := runWithStdout(t, th, cmd.Ps(), []string{"ps"})
		assert.Contains(t, out, "No running processes")
	})

	t.Run("ListsAndFiltersAliveProcess", func(t *testing.T) {
		t.Parallel()

		th := test.SetupCommand(t)
		dag := th.DAG(t, `name: ps-test-dag
steps:
  - name: "1"
    run: echo hello
`)

		runID := "ps-run-1"
		proc, err := th.ProcStore.Acquire(th.Context, dag.ProcGroup(), exec.ProcMeta{
			StartedAt:    time.Now().UnixMilli(),
			Name:         dag.Name,
			DAGRunID:     runID,
			AttemptID:    "attempt-1",
			RootName:     dag.Name,
			RootDAGRunID: runID,
		})
		require.NoError(t, err)
		t.Cleanup(func() {
			_ = proc.Stop(th.Context)
		})

		out := runWithStdout(t, th, cmd.Ps(), []string{"ps"})
		assert.Contains(t, out, dag.Name)
		assert.Contains(t, out, runID)

		out = runWithStdout(t, th, cmd.Ps(), []string{"ps", "-d", dag.Name, "-r", "ps-run"})
		assert.Contains(t, out, dag.Name)
		assert.Contains(t, out, runID)

		out = runWithStdout(t, th, cmd.Ps(), []string{"ps", "-d", "other-dag"})
		assert.Contains(t, out, "No running processes")
	})
}

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

		out := runWithStdout(t, th, cmd.Ls(), []string{"ls"})
		assert.Contains(t, out, "NAME")
		assert.Contains(t, out, "ls-alpha")
		assert.Contains(t, out, "ls-beta")

		out = runWithStdout(t, th, cmd.Ls(), []string{"ls", "alpha"})
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

		out := runWithStdout(t, th, cmd.Ls(), []string{"ls", "-l", "-H", "ls-enriched"})
		assert.Contains(t, out, "LAST_STATUS")
		assert.Contains(t, out, "HISTORY")
		assert.Contains(t, out, "Succeeded")
	})
}

// runWithStdout executes a cobra command and returns captured stdout.
func runWithStdout(t *testing.T, th test.Command, c *cobra.Command, args []string) string {
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
