//go:build unix

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

const signalHelperEnv = "CI_TEST_MAPPING_SIGNAL_HELPER"

type stalledCLITransport struct{}

func (stalledCLITransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, err := fmt.Fprintln(os.Stdout, "request-started"); err != nil {
		return nil, err
	}
	<-req.Context().Done()
	return nil, req.Context().Err()
}

func TestCLISignalHelper(t *testing.T) {
	if os.Getenv(signalHelperEnv) != "1" {
		t.Skip("subprocess helper")
	}
	http.DefaultTransport = stalledCLITransport{}
	os.Args = []string{"ci-test-mapping", "map-verify"}
	Execute()
	t.Fatal("stalled command returned before process termination")
}

func TestCLITerminationSignals(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			child := exec.CommandContext(ctx, executable, "-test.run=^TestCLISignalHelper$")
			child.Env = append(os.Environ(), signalHelperEnv+"=1")
			var stderr bytes.Buffer
			child.Stderr = &stderr
			stdout, err := child.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := child.Start(); err != nil {
				t.Fatal(err)
			}
			finished := make(chan struct{})
			var waitError error
			go func() {
				waitError = child.Wait()
				close(finished)
			}()
			t.Cleanup(func() {
				cancel()
				<-finished
			})
			// Synchronize on the actual command's request before sending a signal.
			started := make(chan string, 1)
			go func() {
				line, _ := bufio.NewReader(stdout).ReadString('\n')
				started <- line
			}()
			select {
			case line := <-started:
				if line != "request-started\n" {
					t.Fatalf("request did not start: %q", line)
				}
			case <-ctx.Done():
				t.Fatal("timed out waiting for command's request")
			}
			if err := child.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case <-finished:
				var exitError *exec.ExitError
				if !errors.As(waitError, &exitError) {
					t.Fatalf("expected signal termination, got %v; stderr: %s", waitError, stderr.String())
				}
				status, ok := exitError.Sys().(syscall.WaitStatus)
				if !ok || !status.Signaled() || status.Signal() != sig {
					t.Fatalf("expected termination by %s, got %v; stderr: %s", sig, exitError.Sys(), stderr.String())
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("command survived %s while its HTTP request was stalled", sig)
			}
		})
	}
}
