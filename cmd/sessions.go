package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wabee-ai/wabee-cli/internal/client"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage agent sessions",
	Long: `View, manage, and export agent sessions.

Sessions track conversations and their execution history.`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent sessions",
	Long: `List recent agent sessions.

Examples:
  wabee sessions list
  wabee sessions list --limit 20
  wabee sessions list --before <session-id>`,
	RunE: runSessionsList,
}

var sessionsGetCmd = &cobra.Command{
	Use:   "get <session-id>",
	Short: "Get session details",
	Long: `Get detailed information about a specific session.

Examples:
  wabee sessions get abc123
  wabee sessions get abc123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionsGet,
}

var sessionsTraceCmd = &cobra.Command{
	Use:   "trace <session-id>",
	Short: "View execution trace",
	Long: `View the structured execution trace for a session.

Examples:
  wabee sessions trace abc123
  wabee sessions trace abc123 --request-id req-001
  wabee sessions trace abc123 --max-content-length 500`,
	Args: cobra.ExactArgs(1),
	RunE: runSessionsTrace,
}

var (
	limitFlag            int
	beforeFlag           string
	requestIDFlag        string
	maxContentLengthFlag int
	fullFlag             bool
)

func init() {
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsGetCmd)
	sessionsCmd.AddCommand(sessionsTraceCmd)

	// List flags
	sessionsListCmd.Flags().IntVar(&limitFlag, "limit", 10, "maximum number of sessions to return")
	sessionsListCmd.Flags().StringVar(&beforeFlag, "before", "", "return sessions before this session ID")

	// Trace flags
	sessionsTraceCmd.Flags().StringVar(&requestIDFlag, "request-id", "", "filter by request ID")
	sessionsTraceCmd.Flags().IntVar(&maxContentLengthFlag, "max-content-length", 200, "maximum content length to display")
	sessionsTraceCmd.Flags().BoolVar(&fullFlag, "full", false, "show full content without truncation")
}

func runSessionsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()
	formatter := getFormatter()

	sessions, err := apiClient.ListSessions(ctx, limitFlag, beforeFlag)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	return formatter.Output(sessions)
}

func runSessionsGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()
	formatter := getFormatter()

	sessionID := args[0]
	session, err := apiClient.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	return formatter.Output(session)
}

func runSessionsTrace(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()
	formatter := getFormatter()

	// Configure content length display
	if fullFlag {
		maxContentLengthFlag = 50000
		formatter.SetMaxContentLength(50000)
	}

	sessionID := args[0]
	trace, err := apiClient.GetSessionTrace(ctx, sessionID, requestIDFlag, maxContentLengthFlag)
	if err != nil {
		return fmt.Errorf("failed to get session trace: %w", err)
	}

	return formatter.Output(trace)
}
