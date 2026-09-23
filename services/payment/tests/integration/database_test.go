//go:build integration

package integration

import (
	"context"
	"database/sql"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"lab2/payment/ent"
	"lab2/payment/internal/storage/postgres"
)

func databaseClients(t *testing.T) (*ent.Client, *ent.Client) {
	t.Helper()
	ctx := t.Context()
	container, err := testcontainers.Run(ctx, "postgres:13-alpine",
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_DB": "test", "POSTGRES_USER": "program", "POSTGRES_PASSWORD": "test",
		}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).
			WithStartupTimeout(time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, container)
	t.Cleanup(func() {
		logs, err := container.Logs(context.Background())
		if err == nil {
			defer func() {
				if err := logs.Close(); err != nil {
					t.Error(err)
				}
			}()
			if t.Failed() {
				buffer, err := io.ReadAll(logs)
				if err != nil {
					t.Error(err)
				}
				t.Log(string(buffer))
			}
		}
	})

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	dsn := "postgres://program:test@" + net.JoinHostPort(host, port.Port()) + "/test?sslmode=disable"
	directory, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	//nolint:gosec // Atlas receives fixed test arguments and the isolated container URL; no shell is used.
	command := exec.CommandContext(ctx, "atlas", "migrate", "apply", "--dir", "file://"+directory, "--url", dsn)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Atlas: %v\n%s", err, output)
	}
	newClient := func() *ent.Client {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(8)
		client := postgres.NewClient(db)
		t.Cleanup(func() {
			if err := client.Close(); err != nil {
				t.Error(err)
			}
		})
		return client
	}
	return newClient(), newClient()
}
