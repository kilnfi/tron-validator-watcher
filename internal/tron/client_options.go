package tron

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type ClientOptionFunc func(*Client) error

func WithBaseURL(baseURL string) ClientOptionFunc {
	return func(c *Client) error {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return fmt.Errorf("failed to parse base URL: %w", err)
		}
		c.baseURL = parsedURL
		return nil
	}
}

func WithHTTPClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *Client) error {
		c.client = httpClient
		return nil
	}
}

func WithLogger(logger *logrus.Logger) ClientOptionFunc {
	return func(c *Client) error {
		c.logger = logger
		return nil
	}
}
