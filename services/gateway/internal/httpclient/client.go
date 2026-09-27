package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"lab2/gateway/internal/models/entities"

	"github.com/samber/oops"
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
	http  *http.Client
	slots chan struct{}
}

func New(cfg Config) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = cfg.MaxIdleConnections
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnectionsPerHost
	transport.IdleConnTimeout = cfg.IdleConnectionTimeout
	return &Client{slots: make(chan struct{}, cfg.MaxInFlight), http: &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			// net/http recognizes this sentinel by identity.
			return http.ErrUseLastResponse
		},
	}}
}

func (c *Client) Call(ctx context.Context, method, address, user string, input, output any) (callErr error) {
	select {
	case <-ctx.Done():
		return oops.Wrap(errors.Join(entities.ErrUpstreamTimeout, ctx.Err()))
	case c.slots <- struct{}{}:
	}
	defer func() { <-c.slots }()
	req, err := newRequest(ctx, method, address, user, input)
	if err != nil {
		return err
	}
	res, err := c.http.Do(req)
	if err != nil {
		failure := entities.ErrUpstreamUnavailable
		var timeout interface{ Timeout() bool }
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &timeout) && timeout.Timeout()) {
			failure = entities.ErrUpstreamTimeout
		}
		return oops.With("method", method).Wrap(errors.Join(failure, err))
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
		return oops.With("upstream_status", res.StatusCode).Wrap(responseError(res.StatusCode))
	}
	if output != nil {
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(output); err != nil {
			return oops.Wrap(errors.Join(entities.ErrUpstreamUnavailable, err))
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return oops.Wrap(errors.Join(entities.ErrUpstreamUnavailable,
				errors.New("unexpected data after upstream JSON"), err))
		}
		return nil
	}
	_, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		return oops.Wrapf(errors.Join(entities.ErrUpstreamUnavailable, err), "read response")
	}
	return nil
}

func responseError(status int) error {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return entities.ErrInvalidInput
	case http.StatusNotFound:
		return entities.ErrNotFound
	default:
		return entities.ErrUpstreamUnavailable
	}
}

func (c *Client) CloseIdleConnections() { c.http.CloseIdleConnections() }
