package models

import (
	"strings"
	"time"
)

// FlexTime is a custom time type that can parse timestamps with or without timezone
type FlexTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for FlexTime
func (ft *FlexTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try RFC3339 first (with timezone)
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try RFC3339Nano (with timezone and nanoseconds)
	t, err = time.Parse(time.RFC3339Nano, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try without timezone (assume UTC)
	t, err = time.Parse("2006-01-02T15:04:05.999999", s)
	if err == nil {
		ft.Time = t.UTC()
		return nil
	}

	// Try without timezone and without fractional seconds
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		ft.Time = t.UTC()
		return nil
	}

	return err
}

// MarshalJSON implements json.Marshaler for FlexTime
func (ft FlexTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + ft.Time.Format(time.RFC3339) + `"`), nil
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat endpoint
type ChatRequest struct {
	Messages         []Message        `json:"messages"`
	DataFilterHints  []DataFilterHint `json:"data_filter_hints,omitempty"`
	AutoMemoryRetain bool             `json:"auto_memory_retain,omitempty"`
	ContextFiles     []string         `json:"context_files,omitempty"`
	Images           []ImageInput     `json:"images,omitempty"`
	Budget           *Budget          `json:"budget,omitempty"`
}

// DataFilterHint represents a filter hint for tools
type DataFilterHint struct {
	ToolName    string `json:"tool_name"`
	FilterType  string `json:"filter_type,omitempty"`
	FilterKey   string `json:"filter_key"`
	FilterValue string `json:"filter_value"`
}

// ImageInput represents an image input
type ImageInput struct {
	Content string `json:"content"`
	Type    string `json:"type"` // "base64" or "file_path"
}

// Budget represents the budget configuration for agent execution
type Budget struct {
	LocalRecursionLimit int `json:"local_recursion_limit"`
}

// NewChatRequest creates a ChatRequest from a simple user message
func NewChatRequest(userMessage string) ChatRequest {
	return ChatRequest{
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
	}
}

// ChatResponse represents a response from the chat endpoint
type ChatResponse struct {
	Output    string `json:"output"`
	SessionID string `json:"session_id"`
	RequestID string `json:"request_id"`
}

// StreamEvent represents a Server-Sent Event from the streaming endpoint
type StreamEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// StreamEventData represents the parsed data from a stream event
type StreamEventData struct {
	// Server fields
	ID           string      `json:"id,omitempty"`
	Token        string      `json:"token,omitempty"`
	AgentStep    string      `json:"agent_step,omitempty"`
	FinishReason string      `json:"finish_reason,omitempty"`
	Timestamp    int64       `json:"timestamp,omitempty"`
	SourceType   string      `json:"source_type,omitempty"`
	SourceID     string      `json:"source_id,omitempty"`
	SourceName   string      `json:"source_name,omitempty"`
	Usage        interface{} `json:"usage,omitempty"`

	// Event type for special events
	Type string `json:"type,omitempty"`

	// Legacy/compatibility fields
	Content   string                 `json:"content,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// IsComplete returns true if this event indicates the stream is complete
func (e *StreamEventData) IsComplete() bool {
	return e.FinishReason == "stop" || e.FinishReason == "error"
}

// GetContent returns the content/token from the event
func (e *StreamEventData) GetContent() string {
	if e.Token != "" {
		return e.Token
	}
	return e.Content
}

// Session represents an agent session
type Session struct {
	SessionID    string                 `json:"session_id"`
	CheckpointID string                 `json:"checkpoint_id,omitempty"`
	ShortName    string                 `json:"short_name,omitempty"`
	CreatedAt    FlexTime               `json:"created_at"`
	UpdatedAt    FlexTime               `json:"updated_at,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SessionList represents a paginated list of sessions
type SessionList struct {
	Sessions []Session `json:"sessions"`
}

// SessionDetail represents detailed session information with events
type SessionDetail struct {
	Session
	Events []SessionEvent `json:"events,omitempty"`
}

// SessionEvent represents an event within a session
type SessionEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp FlexTime               `json:"timestamp"`
	Content   string                 `json:"content,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionTrace represents a structured execution trace for debugging
type ExecutionTrace struct {
	SessionID        string            `json:"session_id"`
	ExecutionSummary *ExecutionSummary `json:"execution_summary,omitempty"`
	Requests         []RequestTrace    `json:"requests"`
	UserMessages     []SessionMessage  `json:"user_messages,omitempty"`
	Errors           []interface{}     `json:"errors,omitempty"`
}

// ExecutionSummary contains overall execution statistics
type ExecutionSummary struct {
	StartedAt      string `json:"started_at,omitempty"`
	CompletedAt    string `json:"completed_at,omitempty"`
	DurationMs     *int   `json:"duration_ms,omitempty"`
	Status         string `json:"status"`
	TotalRequests  int    `json:"total_requests"`
	TotalSteps     int    `json:"total_steps"`
	TotalToolCalls int    `json:"total_tool_calls"`
	TotalEvents    int    `json:"total_events"`
}

// SessionMessage represents a user message in a session
type SessionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// RequestTrace represents a single request's execution trace
type RequestTrace struct {
	RequestID      string      `json:"request_id"`
	AgentID        string      `json:"agent_id,omitempty"`
	Status         string      `json:"status"`
	StartedAt      string      `json:"started_at,omitempty"`
	CompletedAt    string      `json:"completed_at,omitempty"`
	DurationMs     *int        `json:"duration_ms,omitempty"`
	TotalSteps     int         `json:"total_steps"`
	TotalToolCalls int         `json:"total_tool_calls"`
	TotalEvents    int         `json:"total_events"`
	ErrorMessage   string      `json:"error_message,omitempty"`
	Steps          []TraceStep `json:"steps"`
	UserInput      string      `json:"user_input,omitempty"`
}

// TraceStep represents a step in the execution trace
type TraceStep struct {
	StepNumber int    `json:"step_number"`
	StepName   string `json:"step_name"`
	StartedAt  string `json:"started_at,omitempty"`
	EndedAt    string `json:"ended_at,omitempty"`
	DurationMs *int   `json:"duration_ms,omitempty"`
	Type       string `json:"type"`
	Content    string `json:"content,omitempty"`
	EventCount int    `json:"event_count"`
}

// ToolCallInfo represents information about a tool call
type ToolCallInfo struct {
	Name         string                 `json:"name"`
	Input        map[string]interface{} `json:"input,omitempty"`
	Output       interface{}            `json:"output,omitempty"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	DurationMs   *int                   `json:"duration_ms,omitempty"`
}

// ToolInfo represents information about an available tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// AgentInfo represents agent metadata
type AgentInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Tools       []string `json:"tools,omitempty"`
}

// HealthStatus represents agent health check response
type HealthStatus struct {
	Status  string            `json:"status"`
	Version string            `json:"version,omitempty"`
	Uptime  float64           `json:"uptime_seconds,omitempty"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// APIError represents an error response from the API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Message + ": " + e.Detail
	}
	return e.Message
}
