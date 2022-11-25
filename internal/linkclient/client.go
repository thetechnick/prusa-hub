package linkclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	opts       ClientOptions
	httpClient *http.Client
}

// Creates a new OCM client with the given options.
func NewClient(opts ...clientOption) *Client {
	c := &Client{}
	for _, opt := range opts {
		opt.ApplyToClientOptions(&c.opts)
	}

	c.httpClient = &http.Client{}
	return c
}

type APIError struct {
	StatusCode int
}

func (e APIError) Error() string {
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

func (c *Client) do(
	ctx context.Context,
	httpMethod string,
	path string,
	params url.Values,
	payload, result interface{},
) error {
	// Build URL
	reqURL, err := url.Parse(c.opts.Endpoint)
	if err != nil {
		return fmt.Errorf("parsing endpoint URL: %w", err)
	}
	reqURL = reqURL.ResolveReference(&url.URL{
		Path: strings.TrimLeft(path, "/"), // trim first slash to always be relative to baseURL
	})

	// Payload
	var resBody io.Reader
	if payload != nil {
		j, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshaling json: %w", err)
		}

		resBody = bytes.NewBuffer(j)
	}

	var fullUrl string
	if len(params) > 0 {
		fullUrl = reqURL.String() + "?" + params.Encode()
	} else {
		fullUrl = reqURL.String()
	}

	httpReq, err := http.NewRequestWithContext(ctx, httpMethod, fullUrl, resBody)
	if err != nil {
		return fmt.Errorf("creating http request: %w", err)
	}

	// Headers
	if len(c.opts.APIKey) > 0 {
		httpReq.Header.Add("X-Api-Key", c.opts.APIKey)
	}
	httpReq.Header.Add("Content-Type", "application/json")

	httpRes, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("executing http request: %w", err)
	}
	defer httpRes.Body.Close()

	// HTTP Error handling
	if httpRes.StatusCode >= 400 && httpRes.StatusCode <= 599 {
		var ocmErr APIError
		ocmErr.StatusCode = httpRes.StatusCode
		return ocmErr
	}

	// Read response
	if result != nil {
		body, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return fmt.Errorf("reading response body %s: %w", fullUrl, err)
		}

		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("unmarshal json response %s: %w", fullUrl, err)
		}
	}

	return nil
}
