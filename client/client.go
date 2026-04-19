package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Message struct {
	ID        string
	Body      []byte
	LockToken string
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) CreateQueue(ctx context.Context, queueName string, maxDelivery *int) error {
	url := fmt.Sprintf("%s/queues/%s", c.baseURL, queueName)
	if maxDelivery != nil {
		url += fmt.Sprintf("?maxDelivery=%d", *maxDelivery)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	_, err := c.httpClient.Do(req)
	return err
}

func (c *Client) Enqueue(ctx context.Context, queueName string, body []byte) error {
	url := fmt.Sprintf("%s/queues/%s/message", c.baseURL, queueName)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	_, err := c.httpClient.Do(req)

	return err
}

func (c *Client) Peek(ctx context.Context, queueName string) (*Message, error) {
	url := fmt.Sprintf("%s/queues/%s", c.baseURL, queueName)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	msg := &Message{}
	err = json.NewDecoder(resp.Body).Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Client) Receive(ctx context.Context, queueName string) (*Message, error) {
	url := fmt.Sprintf("%s/queues/%s/message", c.baseURL, queueName)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	msg := &Message{}
	err = json.NewDecoder(resp.Body).Decode(msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Client) Ack(ctx context.Context, lockToken string) error {
	url := fmt.Sprintf("%s/ack?lockToken=%s", c.baseURL, lockToken)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	_, err := c.httpClient.Do(req)
	return err
}

func (c *Client) Nack(ctx context.Context, lockToken string) error {
	url := fmt.Sprintf("%s/nack?lockToken=%s", c.baseURL, lockToken)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	_, err := c.httpClient.Do(req)
	return err
}

func (c *Client) DeleteQueue(ctx context.Context, queueName string, force *string) error {
	url := ""
	if force != nil {
		url = fmt.Sprintf("%s/queues/%s?force=%s", c.baseURL, queueName, *force)
	} else {
		url = fmt.Sprintf("%s/queues/%s", c.baseURL, queueName)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	_, err := c.httpClient.Do(req)
	return err
}

// func (c *Client) GetStats(queryName string) ([]Stats, error) {}
