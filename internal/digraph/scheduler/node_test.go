package scheduler_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/dagu-org/dagu/internal/digraph"
	"github.com/dagu-org/dagu/internal/digraph/scheduler"
	"github.com/dagu-org/dagu/internal/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	t.Parallel()

	t.Run("Execute", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCommand("true"))
		node.Execute(t)
	})
	t.Run("Error", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCommand("false"))
		node.ExecuteFail(t, "exit status 1")
	})
	t.Run("Signal", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCommand("sleep 3"))
		go func() {
			time.Sleep(100 * time.Millisecond)
			node.Signal(node.Context, syscall.SIGTERM, false)
		}()

		node.SetStatus(scheduler.NodeStatusRunning)

		node.ExecuteFail(t, "signal: terminated")
		require.Equal(t, scheduler.NodeStatusCancel.String(), node.State().Status.String())
	})
	t.Run("SignalOnStop", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCommand("sleep 3"), withNodeSignalOnStop("SIGINT"))
		go func() {
			time.Sleep(100 * time.Millisecond)
			node.Signal(node.Context, syscall.SIGTERM, true) // allow override signal
		}()

		node.SetStatus(scheduler.NodeStatusRunning)

		node.ExecuteFail(t, "signal: interrupt")
		require.Equal(t, scheduler.NodeStatusCancel.String(), node.State().Status.String())
	})
	t.Run("LogOutput", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCommand("echo hello"))
		node.Execute(t)
		node.AssertLogContains(t, "hello")
	})
	t.Run("Stdout", func(t *testing.T) {
		t.Parallel()

		random := path.Join(os.TempDir(), uuid.Must(uuid.NewV7()).String())
		defer func() {
			_ = os.Remove(random)
		}()

		node := setupNode(t, withNodeCommand("echo hello"), withNodeStdout(random))
		node.Execute(t)

		file := node.NodeData().Step.Stdout
		dat, _ := os.ReadFile(file)
		require.Equalf(t, "hello\n", string(dat), "unexpected stdout content: %s", string(dat))
	})
	t.Run("Stderr", func(t *testing.T) {
		t.Parallel()

		random := path.Join(os.TempDir(), uuid.Must(uuid.NewV7()).String())
		defer func() {
			_ = os.Remove(random)
		}()

		node := setupNode(t,
			withNodeCommand("sh"),
			withNodeStderr(random),
			withNodeScript("echo hello >&2"),
		)
		node.Execute(t)

		file := node.NodeData().Step.Stderr
		dat, _ := os.ReadFile(file)
		require.Equalf(t, "hello\n", string(dat), "unexpected stderr content: %s", string(dat))
	})
	t.Run("Output", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs("echo hello"), withNodeOutput("OUTPUT_TEST"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT_TEST", "hello")
	})
	t.Run("OutputJSON", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo '{"key": "value"}'`), withNodeOutput("OUTPUT_JSON_TEST"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT_JSON_TEST", `{"key": "value"}`)
	})
	t.Run("OutputJSONUnescaped", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo {\"key\":\"value\"}`), withNodeOutput("OUTPUT_JSON_TEST"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT_JSON_TEST", `{"key":"value"}`)
	})
	t.Run("OutputTabWithDoubleQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo "hello\tworld"`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", "hello\tworld")
	})
	t.Run("OutputTabWithMixedQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo hello"\t"world`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", "hello\tworld") // This behavior is aligned with bash
	})
	t.Run("OutputTabWithoutQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo hello\tworld`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `hellotworld`) // This behavior is aligned with bash
	})
	t.Run("OutputNewlineCharacter", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo hello\nworld`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `hellonworld`) // This behavior is aligned with bash
	})
	t.Run("OutputEscapedJSONWithoutQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo {\"key\":\"value\"}`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `{"key":"value"}`)
	})
	t.Run("OutputEscapedJSONWithQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo "{\"key\":\"value\"}"`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `{"key":"value"}`)
	})
	t.Run("OutputSingleQuotedString", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo 'hello world'`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `hello world`)
	})
	t.Run("OutputMixedQuotesWithSpace", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo hello "world"`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `hello world`)
	})
	t.Run("OutputNestedQuotes", func(t *testing.T) {
		t.Parallel()

		node := setupNode(t, withNodeCmdArgs(`echo 'hello "world"'`), withNodeOutput("OUTPUT"))
		node.Execute(t)
		node.AssertOutput(t, "OUTPUT", `hello "world"`)
	})
}

type nodeHelper struct {
	*scheduler.Node
	test.Helper
}

type nodeOption func(*scheduler.NodeData)

func withNodeCmdArgs(cmd string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.CmdWithArgs = cmd
	}
}

func withNodeCommand(command string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.Command = command
	}
}

func withNodeSignalOnStop(signal string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.SignalOnStop = signal
	}
}

func withNodeStdout(stdout string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.Stdout = stdout
	}
}

func withNodeStderr(stderr string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.Stderr = stderr
	}
}

func withNodeScript(script string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.Script = script
	}
}

func withNodeOutput(output string) nodeOption {
	return func(data *scheduler.NodeData) {
		data.Step.Output = output
	}
}

func setupNode(t *testing.T, opts ...nodeOption) nodeHelper {
	t.Helper()

	th := test.Setup(t)

	data := scheduler.NodeData{Step: digraph.Step{}}
	for _, opt := range opts {
		opt(&data)
	}

	return nodeHelper{
		Helper: th,
		Node:   scheduler.NodeWithData(data),
	}
}

func (n nodeHelper) Execute(t *testing.T) {
	t.Helper()

	dagRunID := uuid.Must(uuid.NewV7()).String()
	err := n.Setup(n.Context, n.Config.Paths.LogDir, dagRunID)
	require.NoError(t, err, "failed to setup node")

	err = n.Node.Execute(n.execContext(dagRunID))
	require.NoError(t, err, "failed to execute node")

	err = n.Teardown(n.Context)
	require.NoError(t, err, "failed to teardown node")
}

func (n nodeHelper) ExecuteFail(t *testing.T, expectedErr string) {
	t.Helper()

	dagRunID := uuid.Must(uuid.NewV7()).String()
	err := n.Node.Execute(n.execContext(dagRunID))
	require.Error(t, err, "expected error")
	require.Contains(t, err.Error(), expectedErr, "unexpected error")
}

func (n nodeHelper) AssertLogContains(t *testing.T, expected string) {
	t.Helper()

	dat, err := os.ReadFile(n.StdoutFile())
	require.NoErrorf(t, err, "failed to read log file %q", n.StdoutFile())
	require.Contains(t, string(dat), expected, "log file does not contain expected string")
}

func (n nodeHelper) AssertOutput(t *testing.T, key, value string) {
	t.Helper()

	require.NotNil(t, n.NodeData().State.OutputVariables, "output variables not set")
	data, ok := n.NodeData().State.OutputVariables.Load(key)
	require.True(t, ok, "output variable not found")
	require.Equal(t, fmt.Sprintf(`%s=%s`, key, value), data, "output variable value mismatch")
}

func (n nodeHelper) execContext(dagRunID string) context.Context {
	return digraph.SetupEnv(n.Context, &digraph.DAG{}, nil, digraph.DAGRunRef{}, dagRunID, "logFile", nil)
}
