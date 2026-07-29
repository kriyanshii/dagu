// Copyright (C) 2026 Yota Hamada
// SPDX-License-Identifier: GPL-3.0-or-later

package cmd_test

import (
	"fmt"
	"testing"

	"github.com/dagucloud/dagu/internal/cmd"
	"github.com/dagucloud/dagu/internal/core"
	"github.com/dagucloud/dagu/internal/test"
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
