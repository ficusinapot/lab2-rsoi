package rest

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestServeShutdown(t *testing.T) {
	t.Parallel()
	for _, forced := range []bool{false, true} {
		t.Run(map[bool]string{false: "graceful", true: "deadline"}[forced], func(t *testing.T) {
			t.Parallel()
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := listener.Addr().String()
			if err := listener.Close(); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			entered := make(chan struct{})
			release := make(chan struct{})
			var once sync.Once
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/ready" {
					w.WriteHeader(http.StatusOK)
					return
				}
				once.Do(func() { close(entered) })
				select {
				case <-release:
					w.WriteHeader(http.StatusNoContent)
				case <-r.Context().Done():
				}
			})
			cfg := Config{
				Address: address, ReadHeaderTimeout: time.Second, ReadTimeout: time.Second,
				WriteTimeout: time.Second, IdleTimeout: time.Second, ShutdownTimeout: 50 * time.Millisecond,
			}
			stopped := make(chan error, 1)
			go func() { stopped <- Serve(ctx, cfg, handler) }()
			client := &http.Client{Timeout: time.Second}
			defer client.CloseIdleConnections()
			deadline := time.Now().Add(time.Second)
			for {
				res, err := client.Get("http://" + address + "/ready")
				if err == nil {
					if err := res.Body.Close(); err != nil {
						t.Fatal(err)
					}
					break
				}
				if time.Now().After(deadline) {
					t.Fatal(err)
				}
				time.Sleep(time.Millisecond)
			}
			response := make(chan error, 1)
			go func() {
				res, err := client.Get("http://" + address + "/active")
				if err == nil {
					_, readErr := io.Copy(io.Discard, res.Body)
					err = errors.Join(readErr, res.Body.Close())
				}
				response <- err
			}()
			<-entered
			cancel()
			if !forced {
				close(release)
			}
			shutdownErr := <-stopped
			requestErr := <-response
			if forced {
				if !errors.Is(shutdownErr, context.DeadlineExceeded) || requestErr == nil {
					t.Fatalf("shutdown %v request %v", shutdownErr, requestErr)
				}
			} else if shutdownErr != nil || requestErr != nil {
				t.Fatalf("shutdown %v request %v", shutdownErr, requestErr)
			}
		})
	}
}
