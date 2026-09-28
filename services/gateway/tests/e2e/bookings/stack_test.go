//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

const loyaltyServiceName = "loyalty"

type stack struct {
	gateway, loyalty string
	urls             []string
	db               *sql.DB
	containers       map[string]testcontainers.Container
	addresses        map[string]string
}

func newStack(t *testing.T) *stack {
	t.Helper()
	ctx := context.Background()
	root := os.Getenv("GATEWAY_PROJECT_ROOT")
	if root == "" {
		t.Fatal("GATEWAY_PROJECT_ROOT is required")
	}
	nw, err := network.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := nw.Remove(context.Background()); err != nil {
			t.Error(err)
		}
	})
	dbContainer, err := testcontainers.Run(ctx, "postgres:13-alpine",
		network.WithNetwork([]string{"reservation-db"}, nw),
		testcontainers.WithEnv(map[string]string{
			"POSTGRES_DB": "reservations", "POSTGRES_USER": "program", "POSTGRES_PASSWORD": "test",
		}),
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).
			WithStartupTimeout(2*time.Minute)))
	if err != nil {
		t.Fatal(err)
	}
	testcontainers.CleanupContainer(t, dbContainer)
	host, err := dbContainer.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := dbContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatal(err)
	}
	dsn := "postgres://program:test@" + net.JoinHostPort(host, port.Port()) + "/reservations?sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, name := range []string{"payments", "loyalties"} {
		// Database names are fixed test constants, never request input.
		if _, err := db.ExecContext(ctx, "CREATE DATABASE "+name); err != nil {
			t.Fatal(err)
		}
	}
	out := &stack{db: db, containers: make(map[string]testcontainers.Container), addresses: make(map[string]string)}
	for _, svc := range []struct{ name, port string }{
		{"reservation", "8070/tcp"}, {"payment", "8060/tcp"}, {loyaltyServiceName, "8050/tcp"}, {"gateway", "8080/tcp"},
	} {
		image := testcontainers.ContainerCustomizer(testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:    root,
			Dockerfile: filepath.Join("services", svc.name, "Dockerfile"), KeepImage: true, BuildLogWriter: io.Discard,
			BuildOptionsModifier: func(opts *build.ImageBuildOptions) { opts.Version = build.BuilderBuildKit },
		}))
		if os.Getenv("GATEWAY_E2E_PREBUILT") == "1" {
			image = testcontainers.WithImage("lab2-" + svc.name)
		}
		container, err := testcontainers.Run(ctx, "",
			image,
			network.WithNetwork([]string{svc.name}, nw), testcontainers.WithExposedPorts(svc.port),
			testcontainers.WithWaitStrategy(wait.ForHTTP("/manage/health").WithPort(nat.Port(svc.port)).
				WithStartupTimeout(3*time.Minute)))
		if err != nil {
			t.Fatal(err)
		}
		testcontainers.CleanupContainer(t, container)
		captureLogs(t, container, svc.name)
		host, err := container.Host(ctx)
		if err != nil {
			t.Fatal(err)
		}
		port, err := container.MappedPort(ctx, nat.Port(svc.port))
		if err != nil {
			t.Fatal(err)
		}
		address := "http://" + net.JoinHostPort(host, port.Port())
		out.urls = append(out.urls, address)
		out.containers[svc.name] = container
		out.addresses[svc.name] = address
		if svc.name == "gateway" {
			out.gateway = address
		}
		if svc.name == loyaltyServiceName {
			out.loyalty = address
		}
	}
	_, err = db.ExecContext(ctx, `INSERT INTO hotels (hotel_uid,name,country,city,address,stars,price)
 VALUES ('049161bb-badd-4fa8-9d90-87c9a82b0668','Ararat Park Hyatt Moscow',
 'Россия','Москва','Неглинная ул., 4',5,10000)`)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func captureLogs(t *testing.T, container testcontainers.Container, name string) {
	t.Helper()
	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logs, err := container.Logs(ctx)
		if err != nil {
			t.Logf("%s logs: %v", name, err)
			return
		}
		body, readErr := io.ReadAll(logs)
		if readErr != nil {
			t.Logf("%s read logs: %v", name, readErr)
		}
		if err := logs.Close(); err != nil {
			t.Logf("%s close logs: %v", name, err)
		}
		t.Logf("%s logs:\n%s", name, body)
	})
}
