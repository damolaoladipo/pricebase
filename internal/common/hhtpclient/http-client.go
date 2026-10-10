package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"go.uber.org/zap"
)

// Request describes a single outbound HTTP call. Body is raw bytes so the
// client can replay it across retries without the caller rebuilding a reader.
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// Response is a fully-read HTTP response. Send drains and closes the network
// body before returning, so callers never manage the connection lifecycle.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// ClientConfig tunes the transport. Zero values fall back to the defaults in
// NewClientWithConfig.
type ClientConfig struct {
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxRetries          int
	RetryWaitTime       time.Duration
}

// Client is the outbound HTTP contract. Services depend on this interface
// rather than *DefaultClient so tests can substitute a stub with no network.
type Client interface {
	Send(ctx context.Context, req *Request) (*Response, error)
}

// DefaultClient is the net/http-backed implementation of Client.
type DefaultClient struct {
	client     *http.Client
	maxRetries int
	retryWait  time.Duration
}

// compile-time check that DefaultClient satisfies Client.
var _ Client = (*DefaultClient)(nil)

// NewDefaultClient returns a client with a 30s timeout and no retries.
func NewDefaultClient() *DefaultClient {
	return NewClientWithConfig(ClientConfig{})
}

// NewClientWithConfig builds a client with connection pooling. Each zero-valued
// field in config falls back to a default.
func NewClientWithConfig(config ClientConfig) *DefaultClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 100
	}
	if config.MaxIdleConnsPerHost == 0 {
		config.MaxIdleConnsPerHost = 10
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = 500 * time.Millisecond
	}

	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return &DefaultClient{
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		maxRetries: config.MaxRetries,
		retryWait:  config.RetryWaitTime,
	}
}

// Send performs req and returns the response, retrying transient failures on
// idempotent methods only.
//
// A status of 400 or above is returned as a *utils.ErrorResponse carrying the
// upstream status. The upstream response body is logged but deliberately not
// placed on the error, so provider internals cannot leak through the error
// middleware to an API client.
func (c *DefaultClient) Send(ctx context.Context, req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, utils.NewErrorResponse("upstream request cancelled", http.StatusGatewayTimeout, nil, nil)
			case <-time.After(c.retryWait * time.Duration(attempt)):
			}
		}

		resp, err := c.attempt(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if !c.shouldRetry(req, err) {
			return nil, err
		}

		applogger.APILogger.Warn("outbound request failed, retrying",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)
	}

	return nil, lastErr
}

// attempt performs one request/response cycle. The request is rebuilt each call
// so Body can be replayed on retry.
func (c *DefaultClient) attempt(ctx context.Context, req *Request) (*Response, error) {
	var body io.Reader
	if req.Body != nil {
		body = bytes.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		applogger.APILogger.Error("could not build outbound request",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
			zap.Error(err),
		)
		return nil, utils.NewErrorResponse("invalid outbound request", http.StatusInternalServerError, nil, nil)
	}

	if req.Body != nil {
		httpReq.ContentLength = int64(len(req.Body))
		httpReq.Header.Set("Content-Type", "application/json")
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	started := time.Now()

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, utils.NewErrorResponse("upstream request failed", http.StatusBadGateway, nil, nil)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, utils.NewErrorResponse("could not read upstream response", http.StatusBadGateway, nil, nil)
	}

	headers := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	applogger.APILogger.Info("outbound request completed",
		zap.String("method", req.Method),
		zap.String("url", req.URL),
		zap.Int("status", resp.StatusCode),
		zap.Duration("duration", time.Since(started)),
	)

	if resp.StatusCode >= http.StatusBadRequest {
		// Body goes to the log, not the error, to keep provider payloads out of
		// client-facing responses.
		applogger.APILogger.Error("upstream returned an error status",
			zap.String("method", req.Method),
			zap.String("url", req.URL),
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", respBody),
		)

		message := fmt.Sprintf("upstream responded with status %d", resp.StatusCode)

		return nil, utils.NewErrorResponse(message, resp.StatusCode, nil, nil)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    headers,
	}, nil
}

// shouldRetry reports whether err is worth another attempt.
//
// Only idempotent methods are retried. POST is excluded on purpose: replaying a
// charge or subscription write after an ambiguous timeout can double-bill. Use
// an idempotency key at the call site if a POST needs retry semantics.
func (c *DefaultClient) shouldRetry(req *Request, err error) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions:
	default:
		return false
	}

	// Transport failures arrive here as a 502 ErrorResponse, so the status
	// check below already covers timeouts and connection resets.
	var apiErr *utils.ErrorResponse
	if errors.As(err, &apiErr) {
		return apiErr.Status >= http.StatusInternalServerError ||
			apiErr.Status == http.StatusTooManyRequests
	}

	return false
}
