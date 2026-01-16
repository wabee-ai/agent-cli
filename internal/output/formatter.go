package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/wabee-ai/wabee-cli/internal/config"
	"github.com/wabee-ai/wabee-cli/pkg/models"
)

// Format represents an output format
type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
	FormatTree  Format = "tree"
	FormatText  Format = "text"
)

// Styles for terminal output
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	boldStyle    = lipgloss.NewStyle().Bold(true)

	// Tree styles
	treeNodeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	treeLeafStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	treeBranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// Formatter handles output formatting
type Formatter struct {
	format           Format
	writer           io.Writer
	useColor         bool
	maxContentLength int // 0 means no truncation
}

// NewFormatter creates a new formatter
func NewFormatter(format string) *Formatter {
	f := Format(format)
	if f == "" {
		f = Format(config.GetOutputFormat())
	}

	return &Formatter{
		format:           f,
		writer:           os.Stdout,
		useColor:         config.UseColor(),
		maxContentLength: 60, // default truncation length
	}
}

// SetWriter sets the output writer
func (f *Formatter) SetWriter(w io.Writer) {
	f.writer = w
}

// SetMaxContentLength sets the maximum content length for display (0 for no truncation)
func (f *Formatter) SetMaxContentLength(length int) {
	f.maxContentLength = length
}

// Success prints a success message
func Success(msg string) {
	if config.UseColor() {
		fmt.Println(successStyle.Render("✓ " + msg))
	} else {
		fmt.Println("OK: " + msg)
	}
}

// Error prints an error message
func Error(msg string) {
	if config.UseColor() {
		fmt.Fprintln(os.Stderr, errorStyle.Render("✗ "+msg))
	} else {
		fmt.Fprintln(os.Stderr, "ERROR: "+msg)
	}
}

// Warn prints a warning message
func Warn(msg string) {
	if config.UseColor() {
		fmt.Println(warnStyle.Render("⚠ " + msg))
	} else {
		fmt.Println("WARN: " + msg)
	}
}

// Info prints an info message
func Info(msg string) {
	if config.UseColor() {
		fmt.Println(infoStyle.Render("ℹ " + msg))
	} else {
		fmt.Println("INFO: " + msg)
	}
}

// Dim prints dimmed text
func Dim(msg string) string {
	if config.UseColor() {
		return dimStyle.Render(msg)
	}
	return msg
}

// Bold returns bold text
func Bold(msg string) string {
	if config.UseColor() {
		return boldStyle.Render(msg)
	}
	return msg
}

// JSON outputs data as JSON
func (f *Formatter) JSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Text outputs plain text
func (f *Formatter) Text(text string) {
	fmt.Fprintln(f.writer, text)
}

// Output outputs data in the configured format
func (f *Formatter) Output(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.JSON(data)
	case FormatTable:
		return f.outputAsTable(data)
	case FormatTree:
		return f.outputAsTree(data)
	default:
		return f.outputAsText(data)
	}
}

// outputAsText outputs data as text
func (f *Formatter) outputAsText(data interface{}) error {
	switch v := data.(type) {
	case string:
		f.Text(v)
	case *models.ChatResponse:
		f.Text(v.Output)
	case *models.SessionList:
		fmt.Fprintf(f.writer, "%-36s  %-30s  %s\n", "ID", "NAME", "CREATED")
		for _, s := range v.Sessions {
			fmt.Fprintf(f.writer, "%-36s  %-30s  %s\n", s.SessionID, truncate(s.ShortName, 30), s.CreatedAt.Format(time.RFC3339))
		}
	case *models.SessionDetail:
		fmt.Fprintf(f.writer, "Session: %s\n", v.SessionID)
		if v.Status != "" {
			fmt.Fprintf(f.writer, "Status: %s\n", v.Status)
		}
		fmt.Fprintf(f.writer, "Created: %s\n", v.CreatedAt.Format(time.RFC3339))
		if !v.UpdatedAt.IsZero() {
			fmt.Fprintf(f.writer, "Updated: %s\n", v.UpdatedAt.Format(time.RFC3339))
		}
		if len(v.Events) > 0 {
			fmt.Fprintf(f.writer, "\nEvents:\n")
			for _, e := range v.Events {
				fmt.Fprintf(f.writer, "  [%s] %s: %s\n", e.Timestamp.Format("15:04:05"), e.Type, truncate(e.Content, 80))
			}
		}
	case *models.ExecutionTrace:
		f.outputTrace(v)
	case *models.HealthStatus:
		status := v.Status
		if f.useColor {
			if status == "healthy" || status == "ok" {
				status = successStyle.Render(status)
			} else {
				status = errorStyle.Render(status)
			}
		}
		fmt.Fprintf(f.writer, "Status: %s\n", status)
		if v.Version != "" {
			fmt.Fprintf(f.writer, "Version: %s\n", v.Version)
		}
		if v.Uptime > 0 {
			fmt.Fprintf(f.writer, "Uptime: %.1fs\n", v.Uptime)
		}
		for name, check := range v.Checks {
			fmt.Fprintf(f.writer, "  %s: %s\n", name, check)
		}
	case *models.AgentInfo:
		fmt.Fprintf(f.writer, "Name: %s\n", v.Name)
		fmt.Fprintf(f.writer, "Version: %s\n", v.Version)
		if v.Description != "" {
			fmt.Fprintf(f.writer, "Description: %s\n", v.Description)
		}
		if len(v.Tools) > 0 {
			fmt.Fprintf(f.writer, "Tools: %s\n", strings.Join(v.Tools, ", "))
		}
	case []models.ToolInfo:
		for _, t := range v {
			fmt.Fprintf(f.writer, "%s\n", Bold(t.Name))
			if t.Description != "" {
				fmt.Fprintf(f.writer, "  %s\n", Dim(t.Description))
			}
		}
	default:
		return f.JSON(data)
	}
	return nil
}

// outputAsTable outputs data as a table
func (f *Formatter) outputAsTable(data interface{}) error {
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)

	switch v := data.(type) {
	case *models.SessionList:
		fmt.Fprintln(w, "ID\tNAME\tCREATED")
		for _, s := range v.Sessions {
			fmt.Fprintf(w, "%s\t%s\t%s\n",
				s.SessionID,
				truncate(s.ShortName, 40),
				s.CreatedAt.Format("2006-01-02 15:04"),
			)
		}
	case []models.ToolInfo:
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		for _, t := range v {
			fmt.Fprintf(w, "%s\t%s\n", t.Name, truncate(t.Description, 60))
		}
	default:
		return f.outputAsText(data)
	}

	return w.Flush()
}

// outputAsTree outputs data as a tree
func (f *Formatter) outputAsTree(data interface{}) error {
	switch v := data.(type) {
	case *models.ExecutionTrace:
		f.outputTrace(v)
	default:
		return f.outputAsText(data)
	}
	return nil
}

// outputTrace outputs an execution trace as a tree
func (f *Formatter) outputTrace(trace *models.ExecutionTrace) {
	// Header
	status := "unknown"
	duration := "N/A"
	if trace.ExecutionSummary != nil {
		status = trace.ExecutionSummary.Status
		if trace.ExecutionSummary.DurationMs != nil {
			duration = fmt.Sprintf("%.1fs", float64(*trace.ExecutionSummary.DurationMs)/1000)
		}
	}
	header := fmt.Sprintf("Session: %s | Status: %s | Duration: %s", trace.SessionID, status, duration)

	if f.useColor {
		fmt.Fprintln(f.writer, boldStyle.Render(header))
	} else {
		fmt.Fprintln(f.writer, header)
	}
	fmt.Fprintln(f.writer)

	for _, req := range trace.Requests {
		reqDuration := "N/A"
		if req.DurationMs != nil {
			reqDuration = fmt.Sprintf("%.1fs", float64(*req.DurationMs)/1000)
		}
		reqLine := fmt.Sprintf("Request: %s (%s) - %s", req.RequestID, reqDuration, req.Status)
		if f.useColor {
			fmt.Fprintln(f.writer, treeNodeStyle.Render(reqLine))
		} else {
			fmt.Fprintln(f.writer, reqLine)
		}

		// Show user input if available
		if req.UserInput != "" {
			input := req.UserInput
			if f.maxContentLength > 0 {
				input = truncate(input, f.maxContentLength)
			}
			inputLine := fmt.Sprintf("  Input: %s", input)
			if f.useColor {
				fmt.Fprintln(f.writer, dimStyle.Render(inputLine))
			} else {
				fmt.Fprintln(f.writer, inputLine)
			}
		}

		for i, step := range req.Steps {
			isLast := i == len(req.Steps)-1
			f.printTraceStep(step, "", isLast)
		}
		fmt.Fprintln(f.writer)
	}
}

// printTraceStep prints a trace step
func (f *Formatter) printTraceStep(step models.TraceStep, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Format step line
	stepName := step.StepName
	if stepName == "" {
		stepName = step.Type
	}
	duration := ""
	if step.DurationMs != nil && *step.DurationMs > 0 {
		duration = fmt.Sprintf(" (%.1fs)", float64(*step.DurationMs)/1000)
	}

	stepLine := fmt.Sprintf("%s%s%s%s", prefix, connector, stepName, duration)

	if f.useColor {
		fmt.Fprintln(f.writer, treeBranchStyle.Render(prefix+connector)+treeNodeStyle.Render(stepName)+dimStyle.Render(duration))
	} else {
		fmt.Fprintln(f.writer, stepLine)
	}

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	if step.Content != "" {
		// Print content if present
		content := step.Content
		if f.maxContentLength > 0 && f.maxContentLength < 1000 {
			// Truncate for short display (single line)
			content = truncate(content, f.maxContentLength)
			if f.useColor {
				fmt.Fprintln(f.writer, treeBranchStyle.Render(childPrefix+"└── ")+treeLeafStyle.Render("\""+content+"\""))
			} else {
				fmt.Fprintln(f.writer, childPrefix+"└── \""+content+"\"")
			}
		} else {
			// Multi-line display: render each line separately with proper styling
			lines := splitContentLines(content)
			for i, line := range lines {
				if i == 0 {
					// First line with tree connector
					if f.useColor {
						fmt.Fprintln(f.writer, treeBranchStyle.Render(childPrefix+"└── ")+treeLeafStyle.Render("\""+line))
					} else {
						fmt.Fprintln(f.writer, childPrefix+"└── \""+line)
					}
				} else if i == len(lines)-1 {
					// Last line with closing quote
					if f.useColor {
						fmt.Fprintln(f.writer, treeBranchStyle.Render(childPrefix+"    ")+treeLeafStyle.Render(line+"\""))
					} else {
						fmt.Fprintln(f.writer, childPrefix+"    "+line+"\"")
					}
				} else {
					// Middle lines
					if f.useColor {
						fmt.Fprintln(f.writer, treeBranchStyle.Render(childPrefix+"    ")+treeLeafStyle.Render(line))
					} else {
						fmt.Fprintln(f.writer, childPrefix+"    "+line)
					}
				}
			}
		}
	}
}

// truncate truncates a string to the specified length and flattens newlines
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ") // Flatten newlines for single-line display
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// splitContentLines splits content into lines, filtering empty lines and normalizing whitespace
func splitContentLines(s string) []string {
	// Normalize line endings (handle \r\n and \r)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimSpace(s)

	if !strings.Contains(s, "\n") {
		return []string{s}
	}

	// Split and filter empty lines
	rawLines := strings.Split(s, "\n")
	var result []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line) // Use TrimSpace to handle all Unicode whitespace
		if line != "" {
			result = append(result, line)
		}
	}

	if len(result) == 0 {
		return []string{s} // Fallback if all lines were empty
	}
	return result
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}
