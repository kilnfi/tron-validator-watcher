package tron

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

type Client struct {
	client  *http.Client
	baseURL *url.URL
	logger  *logrus.Logger

	Network NetworkClient
	Account AccountClient
}

func NewClient(options ...ClientOptionFunc) (*Client, error) {
	c := &Client{}

	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(c); err != nil {
			return nil, err
		}
	}

	if c.baseURL == nil {
		return nil, fmt.Errorf("base URL is required")
	}

	if c.client == nil {
		c.client = http.DefaultClient
	}

	if c.logger == nil {
		c.logger = logrus.New()
		c.logger.SetLevel(logrus.InfoLevel)
		c.logger.SetFormatter(&logrus.TextFormatter{
			ForceColors: true,
		})
	}

	c.Network = NewNetworkClient(c)
	c.Account = NewAccountClient(c)

	return c, nil
}
