// Copyright (c) ideaaiplus
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://api.clouding.io"

// Client manages communication with the Clouding.io API.
type Client struct {
	httpClient *http.Client
	apiKey     string
}

// New constructs a Client authenticated with the given API key.
func New(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

// NotFoundError is returned when the API responds with HTTP 404.
type NotFoundError struct {
	ID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("resource %q not found", e.ID)
}

// ServerCreateRequest is the JSON body for POST /v1/servers.
type ServerCreateRequest struct {
	Name      string        `json:"name"`
	Hostname  string        `json:"hostname"`
	FlavorID  string        `json:"flavorId"`
	Volume    VolumeConfig  `json:"volume"`
	AccessCfg *AccessConfig `json:"accessConfiguration,omitempty"`
}

// VolumeConfig describes the boot volume for a new server.
type VolumeConfig struct {
	Source string `json:"source"`
	ID     string `json:"id"`
	SsdGB  int    `json:"ssdGb"`
}

// AccessConfig holds optional access credentials for a new server.
type AccessConfig struct {
	SSHKeyID string `json:"sshKeyId,omitempty"`
}

// ServerUpdateRequest is the JSON body for PUT /v1/servers/{id}.
type ServerUpdateRequest struct {
	Name     string `json:"name,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// Server represents a Clouding.io server as returned by the API.
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Flavor   string `json:"flavor"`
	Image    struct {
		ID string `json:"id"`
	} `json:"image"`
	SSHKeyID string `json:"sshKeyId,omitempty"`
}

// CreateServer calls POST /v1/servers and returns the created Server.
func (c *Client) CreateServer(ctx context.Context, req ServerCreateRequest) (*Server, error) {
	var srv Server
	if err := c.doRequest(ctx, http.MethodPost, "/v1/servers", req, &srv); err != nil {
		return nil, err
	}
	return &srv, nil
}

// GetServer calls GET /v1/servers/{id} and returns the Server.
// Returns *NotFoundError when the API responds with 404.
func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	var srv Server
	if err := c.doRequest(ctx, http.MethodGet, "/v1/servers/"+id, nil, &srv); err != nil {
		return nil, err
	}
	return &srv, nil
}

// UpdateServer calls PUT /v1/servers/{id} and returns the updated Server.
func (c *Client) UpdateServer(ctx context.Context, id string, req ServerUpdateRequest) (*Server, error) {
	var srv Server
	if err := c.doRequest(ctx, http.MethodPut, "/v1/servers/"+id, req, &srv); err != nil {
		return nil, err
	}
	return &srv, nil
}

// DeleteServer calls DELETE /v1/servers/{id}.
// A 404 response is treated as success (idempotent delete).
func (c *Client) DeleteServer(ctx context.Context, id string) error {
	err := c.doRequest(ctx, http.MethodDelete, "/v1/servers/"+id, nil, nil)
	if err != nil {
		if _, ok := err.(*NotFoundError); ok {
			return nil
		}
		return err
	}
	return nil
}

// doRequest is the single internal method that handles all HTTP communication.
func (c *Client) doRequest(ctx context.Context, method, path string, body, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("X-API-KEY", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{ID: path}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("clouding API %s %s: status %d: %s", method, path, resp.StatusCode, string(msg))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}
