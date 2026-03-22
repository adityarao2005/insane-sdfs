//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcnetwork "github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	suiteTimeout         = 3 * time.Minute
	dockerVersionTimeout = 8 * time.Second
	dockerBuildTimeout   = 120 * time.Second
	dockerSetupTimeout   = 15 * time.Second
	dockerClientTimeout  = 45 * time.Second
)

func TestHolePunchingDockerE2E(t *testing.T) {
	if os.Getenv("SDFS_RUN_E2E") != "1" {
		t.Skip("set SDFS_RUN_E2E=1 to run e2e tests")
	}
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available")
	}
	if out, err := runHostCommandWithTimeout(context.Background(), dockerVersionTimeout, "docker", "version"); err != nil {
		t.Skipf("docker not ready in this environment: %v output=%s", err, out)
	}

	ctx, cancel := context.WithTimeout(context.Background(), suiteTimeout)
	defer cancel()
	repoRoot := mustRepoRoot(t)

	repo := "insane-sdfs-e2e"
	tag := fmt.Sprintf("%d", time.Now().UnixNano())
	image := repo + ":" + tag
	token := fmt.Sprintf("tok-%d", time.Now().UnixNano())

	nw, err := tcnetwork.New(ctx, tcnetwork.WithAttachable(), tcnetwork.WithDriver("bridge"))
	if err != nil {
		t.Fatalf("create test network: %v", err)
	}
	defer func() { _ = nw.Remove(context.Background()) }()

	rendezvousCtx, rendezvousCancel := context.WithTimeout(ctx, dockerBuildTimeout)
	rendezvous, err := testcontainers.GenericContainer(rendezvousCtx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    repoRoot,
				Dockerfile: "test/e2e/Dockerfile",
				Repo:       repo,
				Tag:        tag,
				KeepImage:  true,
			},
			Entrypoint: []string{"/app/rendezvous"},
			Networks:   []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"rendezvous"},
			},
			WaitingFor: wait.ForLog("rendezvous listening on").WithStartupTimeout(dockerSetupTimeout),
		},
	})
	rendezvousCancel()
	if err != nil {
		t.Fatalf("start rendezvous container: %v", err)
	}
	defer func() { _ = rendezvous.Terminate(context.Background()) }()

	_ = rendezvous

	var wg sync.WaitGroup
	wg.Add(2)
	errCh := make(chan error, 2)
	launch := func(peerID string) {
		defer wg.Done()
		out, err := runPunchClientContainer(ctx, image, nw.Name, peerID, token)
		if err != nil {
			errCh <- fmt.Errorf("peer %s failed: %w output=%s", peerID, err, out)
			return
		}
		if !strings.Contains(out, "hole-punch-success") {
			errCh <- fmt.Errorf("peer %s missing success marker: %s", peerID, out)
			return
		}
	}

	go launch("peer-a")
	go launch("peer-b")
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}

func runHostCommandWithTimeout(parent context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return string(b), err
	}
	return string(b), nil
}

func runPunchClientContainer(parent context.Context, image, networkName, peerID, token string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, dockerClientTimeout)
	defer cancel()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      image,
			Entrypoint: []string{"/app/punch-client"},
			Networks:   []string{networkName},
			Cmd: []string{
				"-peer-id", peerID,
				"-token", token,
				"-server", "rendezvous:39000",
				"-timeout", "30s",
			},
			WaitingFor: wait.ForExit().WithExitTimeout(dockerClientTimeout),
		},
	})
	if err != nil {
		return "", err
	}
	defer func() { _ = container.Terminate(context.Background()) }()

	logsReader, logErr := container.Logs(context.Background())
	if logErr != nil {
		return "", logErr
	}
	defer logsReader.Close()
	logs, readErr := io.ReadAll(logsReader)
	if readErr != nil {
		return "", readErr
	}

	state, stateErr := container.State(context.Background())
	if stateErr != nil {
		return string(logs), stateErr
	}
	if state.ExitCode != 0 {
		return string(logs), fmt.Errorf("client exited with code %d", state.ExitCode)
	}

	return string(logs), nil
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine current file path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../.."))
	return root
}
