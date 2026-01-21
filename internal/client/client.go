package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wabee-ai/wabee-cli/internal/config"
	"github.com/wabee-ai/wabee-cli/pkg/models"
)

// apiPrefix is the base path for all API endpoints
const apiPrefix = "/core/v1"

// Client represents the Wabee API client
type Client struct {
	baseURL    string
	apiKey     string
	authToken  string
	httpClient *http.Client
}

// New creates a new API client
func New() *Client {
	timeout := config.GetTimeout()
	if timeout == 0 {
		timeout = 60
	}

	return &Client{
		baseURL:   config.GetEndpoint(),
		apiKey:    config.GetAPIKey(),
		authToken: config.GetAuthToken(),
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// NewWithConfig creates a new API client with explicit configuration
func NewWithConfig(endpoint, apiKey, authToken string, timeout int) *Client {
	if timeout == 0 {
		timeout = 60
	}

	return &Client{
		baseURL:   endpoint,
		apiKey:    apiKey,
		authToken: authToken,
		httpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		},
	}
}

// BaseURL returns the configured base URL
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Type aliases for external usage
type (
	ExecutionTrace = models.ExecutionTrace
	RequestTrace   = models.RequestTrace
	TraceStep      = models.TraceStep
	SessionDetail  = models.SessionDetail
)

// setHeaders sets common headers on a request
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.apiKey != "" {
		req.Header.Set("x-wabee-access", c.apiKey)
	}
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}
}

// doRequest performs an HTTP request and returns the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleResponse processes an HTTP response
func handleResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		apiErr := &models.APIError{
			Code:    resp.StatusCode,
			Message: http.StatusText(resp.StatusCode),
		}
		// Try to parse error response (ignore error, we have defaults)
		_ = json.Unmarshal(body, apiErr)
		return apiErr
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// StreamCallback is called for each event in a stream
type StreamCallback func(event models.StreamEventData) error

// Chat sends a chat message and returns the response
func (c *Client) Chat(ctx context.Context, input string, sessionID string, budget *int) (*models.ChatResponse, error) {
	req := models.NewChatRequest(input)

	// Add budget if specified
	if budget != nil {
		req.Budget = &models.Budget{
			LocalRecursionLimit: *budget,
		}
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+apiPrefix+"/chain", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)
	// Set session_id and request_id as headers
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	httpReq.Header.Set("session-id", sessionID)
	httpReq.Header.Set("request-id", uuid.New().String())

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result models.ChatResponse
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// StreamResult contains metadata from a streaming request
type StreamResult struct {
	SessionID string
	RequestID string
}

// ChatStream sends a chat message and streams the response
func (c *Client) ChatStream(ctx context.Context, input string, sessionID string, budget *int, callback StreamCallback) (*StreamResult, error) {
	req := models.NewChatRequest(input)

	// Add budget if specified
	if budget != nil {
		req.Budget = &models.Budget{
			LocalRecursionLimit: *budget,
		}
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+apiPrefix+"/chain_streaming", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")
	// Set session_id and request_id as headers
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	requestID := uuid.New().String()
	httpReq.Header.Set("session-id", sessionID)
	httpReq.Header.Set("request-id", requestID)

	result := &StreamResult{
		SessionID: sessionID,
		RequestID: requestID,
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", c.baseURL+apiPrefix+"/chain_streaming", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		apiErr := &models.APIError{
			Code:    resp.StatusCode,
			Message: http.StatusText(resp.StatusCode),
		}
		if err := json.Unmarshal(body, apiErr); err != nil || apiErr.Detail == "" {
			// If we can't parse as APIError or no detail, include raw body
			apiErr.Detail = string(body)
		}
		return nil, apiErr
	}

	// Check Content-Type to ensure we're getting an SSE stream
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected content-type %q (expected text/event-stream), body: %s", contentType, string(body))
	}

	if err := c.parseSSE(resp.Body, callback); err != nil {
		return nil, err
	}

	return result, nil
}

// parseSSE parses Server-Sent Events from a reader
func (c *Client) parseSSE(reader io.Reader, callback StreamCallback) error {
	scanner := bufio.NewScanner(reader)
	var eventType string
	var dataBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line indicates end of event
			if dataBuilder.Len() > 0 {
				data := dataBuilder.String()
				var eventData models.StreamEventData

				if err := json.Unmarshal([]byte(data), &eventData); err != nil {
					// If not JSON, treat as plain text content
					eventData = models.StreamEventData{
						Type:    eventType,
						Content: data,
					}
				}

				if eventType != "" {
					eventData.Type = eventType
				}

				if err := callback(eventData); err != nil {
					return err
				}

				eventType = ""
				dataBuilder.Reset()
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)
			if dataBuilder.Len() > 0 {
				dataBuilder.WriteString("\n")
			}
			dataBuilder.WriteString(data)
		}
	}

	// Handle any remaining data
	if dataBuilder.Len() > 0 {
		data := dataBuilder.String()
		var eventData models.StreamEventData
		if err := json.Unmarshal([]byte(data), &eventData); err != nil {
			eventData = models.StreamEventData{
				Type:    eventType,
				Content: data,
			}
		}
		if eventType != "" {
			eventData.Type = eventType
		}
		if err := callback(eventData); err != nil {
			return err
		}
	}

	return scanner.Err()
}

// ListSessions returns a list of sessions
func (c *Client) ListSessions(ctx context.Context, limit int, before string) (*models.SessionList, error) {
	path := apiPrefix + "/sessions"
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if before != "" {
		params.Set("before", before)
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.SessionList
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSession returns session details
func (c *Client) GetSession(ctx context.Context, sessionID string) (*models.SessionDetail, error) {
	resp, err := c.doRequest(ctx, "GET", apiPrefix+"/sessions/"+sessionID, nil)
	if err != nil {
		return nil, err
	}

	var result models.SessionDetail
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSessionTrace returns the execution trace for a session
func (c *Client) GetSessionTrace(ctx context.Context, sessionID string, requestID string, maxContentLength int) (*models.ExecutionTrace, error) {
	path := apiPrefix + "/sessions/" + sessionID + "/trace"
	params := url.Values{}
	if requestID != "" {
		params.Set("request_id", requestID)
	}
	if maxContentLength > 0 {
		params.Set("max_content_length", fmt.Sprintf("%d", maxContentLength))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var result models.ExecutionTrace
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteSession deletes a session
func (c *Client) DeleteSession(ctx context.Context, sessionID string) error {
	resp, err := c.doRequest(ctx, "DELETE", apiPrefix+"/sessions/"+sessionID, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, nil)
}

// GetTools returns the list of available tools
func (c *Client) GetTools(ctx context.Context) ([]models.ToolInfo, error) {
	resp, err := c.doRequest(ctx, "GET", apiPrefix+"/tools", nil)
	if err != nil {
		return nil, err
	}

	var result []models.ToolInfo
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetAgentInfo returns agent metadata
func (c *Client) GetAgentInfo(ctx context.Context) (*models.AgentInfo, error) {
	resp, err := c.doRequest(ctx, "GET", apiPrefix+"/metadata", nil)
	if err != nil {
		return nil, err
	}

	var result models.AgentInfo
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetHealth returns agent health status
func (c *Client) GetHealth(ctx context.Context) (*models.HealthStatus, error) {
	resp, err := c.doRequest(ctx, "GET", apiPrefix+"/health", nil)
	if err != nil {
		return nil, err
	}

	var result models.HealthStatus
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetActiveStreams returns active streaming sessions
func (c *Client) GetActiveStreams(ctx context.Context) ([]models.Session, error) {
	resp, err := c.doRequest(ctx, "GET", apiPrefix+"/sessions/streams/active", nil)
	if err != nil {
		return nil, err
	}

	var result []models.Session
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
