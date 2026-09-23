package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/samber/oops"
	"lab2/gateway/internal/domain"
)

type Config struct {
	ReservationURL            string        `mapstructure:"reservation_url"`
	PaymentURL                string        `mapstructure:"payment_url"`
	LoyaltyURL                string        `mapstructure:"loyalty_url"`
	Timeout                   time.Duration `mapstructure:"timeout"`
	MaxInFlight               int           `mapstructure:"max_in_flight"`
	MaxIdleConnections        int           `mapstructure:"max_idle_connections"`
	MaxIdleConnectionsPerHost int           `mapstructure:"max_idle_connections_per_host"`
	IdleConnectionTimeout     time.Duration `mapstructure:"idle_connection_timeout"`
}

func (c Config) Validate() error {
	for _, address := range []string{c.ReservationURL, c.PaymentURL, c.LoyaltyURL} {
		u, err := url.Parse(address)
		if err != nil {
			return oops.Wrapf(err, "invalid service URL")
		}
		if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.RawQuery != "" || u.Fragment != "" {
			return oops.Errorf("invalid service URL")
		}
	}
	if c.Timeout <= 0 || c.MaxInFlight <= 0 || c.MaxIdleConnections <= 0 ||
		c.MaxIdleConnectionsPerHost <= 0 || c.IdleConnectionTimeout <= 0 {
		return oops.Errorf("client timeout and connection limits must be positive")
	}
	return nil
}

type Client struct {
	cfg   Config
	http  *http.Client
	slots chan struct{}
}

func New(cfg Config) *Client {
	cfg.ReservationURL = strings.TrimRight(cfg.ReservationURL, "/")
	cfg.PaymentURL = strings.TrimRight(cfg.PaymentURL, "/")
	cfg.LoyaltyURL = strings.TrimRight(cfg.LoyaltyURL, "/")
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = cfg.MaxIdleConnections
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnectionsPerHost
	transport.IdleConnTimeout = cfg.IdleConnectionTimeout
	return &Client{cfg: cfg, slots: make(chan struct{}, cfg.MaxInFlight), http: &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			// net/http recognizes this sentinel by identity.
			return http.ErrUseLastResponse
		},
	}}
}

func (c *Client) call(ctx context.Context, method, address, user string, input, output any) (callErr error) {
	select {
	case <-ctx.Done():
		return oops.Wrap(errors.Join(&domain.UpstreamError{Status: http.StatusGatewayTimeout}, ctx.Err()))
	case c.slots <- struct{}{}:
	}
	defer func() { <-c.slots }()
	req, err := newRequest(ctx, method, address, user, input)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		status := http.StatusBadGateway
		var timeout interface{ Timeout() bool }
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
			status = http.StatusGatewayTimeout
		}
		return oops.With("method", method).Wrap(errors.Join(&domain.UpstreamError{Status: status}, err))
	}
	defer func() { callErr = oops.Wrap(errors.Join(callErr, res.Body.Close())) }()
	return readResponse(res, output)
}

func newRequest(ctx context.Context, method, address, user string, input any) (*http.Request, error) {
	var body bytes.Buffer
	if input != nil {
		if err := json.NewEncoder(&body).Encode(input); err != nil {
			return nil, oops.Wrapf(err, "encode request")
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, address, &body)
	if err != nil {
		return nil, oops.Wrapf(err, "create request")
	}
	if user != "" {
		req.Header.Set("X-User-Name", user)
	}
	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func readResponse(res *http.Response, output any) error {
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return oops.With("upstream_status", res.StatusCode).Wrap(&domain.UpstreamError{Status: responseStatus(res.StatusCode)})
	}
	if output != nil {
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(output); err != nil {
			return oops.Wrap(errors.Join(&domain.UpstreamError{Status: http.StatusBadGateway}, err))
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return oops.Wrap(errors.Join(&domain.UpstreamError{Status: http.StatusBadGateway},
				fmt.Errorf("unexpected data after upstream JSON: %v", err)))
		}
		return nil
	}
	_, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		return oops.Wrapf(errors.Join(&domain.UpstreamError{Status: http.StatusBadGateway}, err), "read response")
	}
	return nil
}

func responseStatus(status int) int {
	if status == http.StatusUnprocessableEntity {
		return http.StatusBadRequest
	}
	if status == http.StatusBadRequest || status == http.StatusNotFound {
		return status
	}
	return http.StatusBadGateway
}

func (c *Client) CloseIdleConnections() { c.http.CloseIdleConnections() }
