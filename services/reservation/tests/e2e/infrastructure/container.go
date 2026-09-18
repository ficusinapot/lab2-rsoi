//go:build e2e

package infrastructure

import (
	"context"
	"net"
	"net/url"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
)

const (
	postgresPort = "5432/tcp"
	appPort      = "8070/tcp"
)

type ReservationStack struct {
	BaseURL     string
	DatabaseURL string
}

func NewReservationStack(t *testing.T) *ReservationStack {
	t.Helper()
	ctx := context.Background()
	net := newNetwork(t, ctx)
	t.Cleanup(func() {
		if err := net.Remove(context.Background()); err != nil {
			t.Error(err)
		}
	})
	_, databaseURL := newPostgres(t, ctx, net)
	_, baseURL := newReservation(t, ctx, net, projectRoot(t))
	return &ReservationStack{BaseURL: baseURL, DatabaseURL: databaseURL}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("RESERVATION_PROJECT_ROOT")
	if root == "" {
		t.Fatal("RESERVATION_PROJECT_ROOT is not set")
	}
	return root
}

func newNetwork(t *testing.T, ctx context.Context) *testcontainers.DockerNetwork {
	t.Helper()
	net, err := network.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return net
}

func databaseURL(host, port string) string {
	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("program", "test"),
		Host:     net.JoinHostPort(host, port),
		Path:     "/reservations",
		RawQuery: url.Values{"sslmode": {"disable"}}.Encode(),
	}).String()
}
