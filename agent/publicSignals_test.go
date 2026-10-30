package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openziti/zrok/v2/agent/agentGrpc"
	"github.com/openziti/zrok/v2/agent/proctree"
	"github.com/openziti/zrok/v2/cmd/zrok2/subordinate"
	"github.com/stretchr/testify/require"
)

func TestPublicSignalWorkerHelper(t *testing.T) {
	if os.Getenv("ZROK_PUBLIC_SIGNAL_WORKER") != "1" {
		return
	}
	for index, arg := range os.Args {
		if arg != "--" || index+2 >= len(os.Args) {
			continue
		}
		ctx, stop, err := subordinate.WorkerContext(context.Background(), os.Stdin)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(os.Args[index+2], []byte(strconv.Itoa(os.Getpid())), 0600))
		fmt.Printf("{\"msg\":\"boot\",\"token\":%q}\n", os.Args[index+1])
		<-ctx.Done()
		stop()
		os.Exit(0)
	}
	t.Fatal("missing worker arguments")
}

func TestPublicWorkerAndAgentStopOrderingsJoinDeletion(t *testing.T) {
	for _, order := range []string{"agent-first", "worker-first"} {
		t.Run(order, func(t *testing.T) {
			if order == "worker-first" && runtime.GOOS == "windows" {
				t.Skip("Windows Process.Signal cannot deliver terminal SIGINT; the agent-first EOF path is portable")
			}
			t.Setenv("ZROK_PUBLIC_SIGNAL_WORKER", "1")
			a := newLifecycleAgent(t)
			pidPath := filepath.Join(t.TempDir(), "worker.pid")
			a.deps.startPublic = func(options proctree.ControlledOptions, args ...string) (*proctree.Child, error) {
				return proctree.StartControlledChild(options, os.Args[0], "-test.run=^TestPublicSignalWorkerHelper$", "--", args[5], pidPath)
			}
			deleting, finishDelete := make(chan struct{}), make(chan struct{})
			var unblock sync.Once
			defer unblock.Do(func() { close(finishDelete) })
			var child *proctree.Child
			var deletes atomic.Int32
			a.deps.deletePublic = func(context.Context, string) error {
				select {
				case <-child.Done():
				default:
					t.Error("deletion preceded worker exit")
				}
				deletes.Add(1)
				close(deleting)
				<-finishDelete
				return nil
			}
			token, _, err := a.SharePublic(publicRequest())
			require.NoError(t, err)
			op := onlyOperation(t, a)
			a.mu.Lock()
			child = op.child
			a.mu.Unlock()
			if order == "worker-first" {
				data, err := os.ReadFile(pidPath)
				require.NoError(t, err)
				pid, err := strconv.Atoi(string(data))
				require.NoError(t, err)
				worker, err := os.FindProcess(pid)
				require.NoError(t, err)
				defer worker.Release()
				require.NoError(t, worker.Signal(os.Interrupt))
				waitLifecycle(t, deleting)
			}
			stopped := asyncResult(a.Shutdown)
			waitLifecycle(t, a.ctx.Done())
			waitLifecycle(t, deleting)
			status, err := (&agentGrpcImpl{agent: a}).Status(context.Background(), &agentGrpc.StatusRequest{})
			require.NoError(t, err)
			require.Len(t, status.Shares, 1)
			require.Equal(t, token, status.Shares[0].Token)
			require.Equal(t, "stopping", status.Shares[0].Status)
			select {
			case <-stopped:
				t.Fatal("shutdown returned before controller deletion")
			default:
			}
			unblock.Do(func() { close(finishDelete) })
			require.NoError(t, lifecycleResult(t, stopped))
			require.NoError(t, proctree.WaitChild(child), "worker should handle its stop rather than require a forced kill")
			require.EqualValues(t, 1, deletes.Load())
			require.Len(t, savedRegistry(t, a).PublicShares, 1)
		})
	}
}
