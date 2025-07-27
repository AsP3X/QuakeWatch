package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"quakewatch-scraper/internal/utils"
)

// EnhancedAPIClient provides enhanced API functionality with circuit breaker and retry logic
type EnhancedAPIClient struct {
	httpClient     *http.Client
	circuitBreaker *utils.CircuitBreaker
	retryStrategy  *utils.RetryStrategy
	metrics        *utils.CollectionMetrics
	logger         *utils.StructuredLogger
	rateLimiter    *utils.RateLimiter
}

// NewEnhancedAPIClient creates a new enhanced API client
func NewEnhancedAPIClient(
	timeout time.Duration,
	circuitBreakerThreshold int,
	circuitBreakerTimeout time.Duration,
	retryStrategy *utils.RetryStrategy,
	metrics *utils.CollectionMetrics,
	logger *utils.StructuredLogger,
	rateLimit int,
) *EnhancedAPIClient {
	return &EnhancedAPIClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		circuitBreaker: utils.NewCircuitBreaker(circuitBreakerThreshold, circuitBreakerTimeout, 3),
		retryStrategy:  retryStrategy,
		metrics:        metrics,
		logger:         logger,
		rateLimiter:    utils.NewRateLimiter(rateLimit, time.Minute),
	}
}

// GetWithRetry performs an HTTP GET request with retry logic and circuit breaker
func (c *EnhancedAPIClient) GetWithRetry(ctx context.Context, url string) (*http.Response, error) {
	start := time.Now()

	// Check rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, utils.NewCollectionError(
			utils.ErrorTypeRateLimit,
			"api_client",
			"rate limit exceeded",
			true,
			err,
		)
	}

	var response *http.Response
	err := c.circuitBreaker.Execute(ctx, func() error {
		return utils.RetryWithBackoff(ctx, c.retryStrategy, func() error {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return utils.NewCollectionError(
					utils.ErrorTypeNetwork,
					"api_client",
					"failed to create request",
					false,
					err,
				)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return utils.NewCollectionError(
					utils.ErrorTypeNetwork,
					"api_client",
					"request failed",
					true,
					err,
				)
			}

			if resp.StatusCode >= 400 {
				err := fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
				if resp.StatusCode >= 500 {
					// Server errors are retryable
					return utils.NewCollectionError(
						utils.ErrorTypeAPI,
						"api_client",
						err.Error(),
						true,
						err,
					)
				} else {
					// Client errors are not retryable
					return utils.NewCollectionError(
						utils.ErrorTypeAPI,
						"api_client",
						err.Error(),
						false,
						err,
					)
				}
			}

			response = resp
			return nil
		})
	})

	duration := time.Since(start)

	// Log the request
	c.logger.LogAPIRequest(ctx, "GET", url, duration, 0, err)

	// Record metrics
	if err != nil {
		if collectionErr, ok := err.(*utils.CollectionError); ok {
			c.metrics.RecordAPIError("api_client", collectionErr.Type.String())
		}
	} else {
		c.metrics.RecordCollectionEnd("api_client", duration, 0, 0)
	}

	return response, err
}

// PostWithRetry performs an HTTP POST request with retry logic and circuit breaker
func (c *EnhancedAPIClient) PostWithRetry(ctx context.Context, url string, body []byte) (*http.Response, error) {
	start := time.Now()

	// Check rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, utils.NewCollectionError(
			utils.ErrorTypeRateLimit,
			"api_client",
			"rate limit exceeded",
			true,
			err,
		)
	}

	var response *http.Response
	err := c.circuitBreaker.Execute(ctx, func() error {
		return utils.RetryWithBackoff(ctx, c.retryStrategy, func() error {
			req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
			if err != nil {
				return utils.NewCollectionError(
					utils.ErrorTypeNetwork,
					"api_client",
					"failed to create request",
					false,
					err,
				)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return utils.NewCollectionError(
					utils.ErrorTypeNetwork,
					"api_client",
					"request failed",
					true,
					err,
				)
			}

			if resp.StatusCode >= 400 {
				err := fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
				if resp.StatusCode >= 500 {
					return utils.NewCollectionError(
						utils.ErrorTypeAPI,
						"api_client",
						err.Error(),
						true,
						err,
					)
				} else {
					return utils.NewCollectionError(
						utils.ErrorTypeAPI,
						"api_client",
						err.Error(),
						false,
						err,
					)
				}
			}

			response = resp
			return nil
		})
	})

	duration := time.Since(start)

	// Log the request
	c.logger.LogAPIRequest(ctx, "POST", url, duration, 0, err)

	// Record metrics
	if err != nil {
		if collectionErr, ok := err.(*utils.CollectionError); ok {
			c.metrics.RecordAPIError("api_client", collectionErr.Type.String())
		}
	} else {
		c.metrics.RecordCollectionEnd("api_client", duration, 0, 0)
	}

	return response, err
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (c *EnhancedAPIClient) GetCircuitBreakerStats() map[string]interface{} {
	return c.circuitBreaker.GetStats()
}

// ResetCircuitBreaker resets the circuit breaker
func (c *EnhancedAPIClient) ResetCircuitBreaker() {
	c.circuitBreaker.Reset()
}
