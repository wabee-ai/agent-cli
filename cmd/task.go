package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wabee-ai/wabee-cli/internal/client"
	"github.com/wabee-ai/wabee-cli/internal/output"
	"github.com/wabee-ai/wabee-cli/pkg/models"
)

var (
	taskSessionFlag string
	taskStreamFlag  bool
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Send tasks to the agent",
	Long: `Send tasks to a Wabee agent and receive responses.

Use 'task new' to start a new task or 'task followup' to continue
an existing session.

Examples:
  # Start a new task
  wabee task new "What is the weather in NYC?"

  # Continue an existing session
  wabee task followup --session abc123 "And tomorrow?"`,
}

var taskNewCmd = &cobra.Command{
	Use:   "new [message]",
	Short: "Start a new task with the agent",
	Long: `Start a new task with a Wabee agent.

The message can be provided as an argument or piped from stdin.

Examples:
  # Simple message
  wabee task new "What is the weather in NYC?"

  # Stream response in real-time
  wabee task new --stream "Explain quantum computing"

  # Pipe input from file
  cat prompt.txt | wabee task new

  # Output as JSON
  wabee task new --output json "List 5 items"`,
	RunE: runTaskNew,
}

var taskFollowupCmd = &cobra.Command{
	Use:   "followup --session <session-id> [message]",
	Short: "Continue a task in an existing session",
	Long: `Send a follow-up message to continue a task in an existing session.

The session ID is required to continue an existing conversation.

Examples:
  # Continue a conversation
  wabee task followup --session abc123 "And tomorrow?"

  # Stream the response
  wabee task followup --session abc123 --stream "Tell me more"

  # Pipe input from file
  cat followup.txt | wabee task followup --session abc123`,
	RunE: runTaskFollowup,
}

func init() {
	// Flags for 'task new'
	taskNewCmd.Flags().BoolVar(&taskStreamFlag, "stream", true, "stream response in real-time")

	// Flags for 'task followup'
	taskFollowupCmd.Flags().StringVarP(&taskSessionFlag, "session", "s", "", "session ID to continue conversation (required)")
	_ = taskFollowupCmd.MarkFlagRequired("session")
	taskFollowupCmd.Flags().BoolVar(&taskStreamFlag, "stream", true, "stream response in real-time")

	// Add subcommands to task
	taskCmd.AddCommand(taskNewCmd)
	taskCmd.AddCommand(taskFollowupCmd)
}

func runTaskNew(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()

	// Get message from args or stdin
	message, err := getTaskMessage(args)
	if err != nil {
		return err
	}

	if message == "" {
		return fmt.Errorf("no message provided. Use: wabee task new \"your message\" or pipe from stdin")
	}

	// Determine output format
	format := getOutputFormat()

	// Non-streaming mode or JSON output
	if !taskStreamFlag || format == "json" {
		return runNonStreamingTask(ctx, apiClient, message, "")
	}

	// Streaming mode
	return runStreamingTask(ctx, apiClient, message, "")
}

func runTaskFollowup(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()

	// Get message from args or stdin
	message, err := getTaskMessage(args)
	if err != nil {
		return err
	}

	if message == "" {
		return fmt.Errorf("no message provided. Use: wabee task followup --session <id> \"your message\" or pipe from stdin")
	}

	// Determine output format
	format := getOutputFormat()

	// Non-streaming mode or JSON output
	if !taskStreamFlag || format == "json" {
		return runNonStreamingTask(ctx, apiClient, message, taskSessionFlag)
	}

	// Streaming mode
	return runStreamingTask(ctx, apiClient, message, taskSessionFlag)
}

func getTaskMessage(args []string) (string, error) {
	// Check if message is provided as argument
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	// Check if stdin has data
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin has data
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	return "", nil
}

func runNonStreamingTask(ctx context.Context, apiClient *client.Client, message, sessionID string) error {
	formatter := getFormatter()
	format := getOutputFormat()

	if !isQuiet() && isVerbose() {
		if sessionID != "" {
			output.Info(fmt.Sprintf("Sending message to session %s...", sessionID))
		} else {
			output.Info("Starting new task...")
		}
	}

	resp, err := apiClient.Chat(ctx, message, sessionID)
	if err != nil {
		return fmt.Errorf("task failed: %w", err)
	}

	// Print answer label for text output
	if format != "json" {
		fmt.Println(output.Bold("Answer:"))
	}

	if err := formatter.Output(resp); err != nil {
		return err
	}

	// Show session and request IDs for text output
	if format != "json" && resp.SessionID != "" && !isQuiet() {
		fmt.Println()
		output.Info(fmt.Sprintf("Session ID: %s", resp.SessionID))
		output.Info(fmt.Sprintf("Request ID: %s", resp.RequestID))
	}

	return nil
}

func runStreamingTask(ctx context.Context, apiClient *client.Client, message, sessionID string) error {
	var responseBuilder strings.Builder
	var spinner *output.Spinner
	answerLabelPrinted := false

	// Show spinner while waiting
	if !isQuiet() {
		spinner = output.NewSpinner("Thinking...")
	}

	result, err := apiClient.ChatStream(ctx, message, sessionID, func(event models.StreamEventData) error {
		// Check for errors
		if event.FinishReason == "error" {
			if spinner != nil {
				spinner.Stop()
				spinner = nil
			}
			return fmt.Errorf("agent error: %s", event.GetContent())
		}

		// Handle based on agent_step
		switch event.AgentStep {
		case "TYPING_TEXT", "FINAL_ANSWER":
			// Stop spinner on first content
			if spinner != nil {
				spinner.Stop()
				spinner = nil
			}

			// Print answer label before first content
			if !answerLabelPrinted {
				fmt.Println(output.Bold("Answer:"))
				answerLabelPrinted = true
			}

			// Print content as it arrives
			content := event.GetContent()
			if content != "" {
				fmt.Print(content)
				responseBuilder.WriteString(content)
			}

		case "PLANNING", "ASSESSING_COMPLEXITY":
			if spinner != nil {
				spinner.UpdateMessage("Planning...")
			}

		case "REASONING", "ADAPTING_APPROACH":
			if spinner != nil {
				spinner.UpdateMessage("Reasoning...")
			}

		case "SELECTING_TOOL":
			if spinner != nil {
				spinner.UpdateMessage("Selecting tool...")
			}

		case "TOOL_CALLING":
			if spinner != nil {
				spinner.UpdateMessage("Executing tool...")
			}

		case "TOOL_FEEDBACK":
			if spinner != nil {
				spinner.UpdateMessage("Processing tool output...")
			}

		case "DELEGATING", "TASK_DELEGATION":
			if spinner != nil {
				spinner.UpdateMessage("Delegating to sub-agent...")
			}

		case "STARTING", "PREPARING":
			if spinner != nil {
				spinner.UpdateMessage("Starting...")
			}

		case "END":
			if spinner != nil {
				spinner.Stop()
				spinner = nil
			}
		}

		return nil
	})

	// Always stop spinner after stream ends (in case no clearing event was received)
	if spinner != nil {
		spinner.Stop()
		spinner = nil
	}

	// Ensure newline after response
	if responseBuilder.Len() > 0 {
		fmt.Println()
	}

	if err != nil {
		return err
	}

	// If no content was received and no error, inform the user
	if responseBuilder.Len() == 0 {
		return fmt.Errorf("no response received from agent")
	}

	// Always show session_id and request_id after the response
	if result != nil && !isQuiet() {
		fmt.Println()
		output.Info(fmt.Sprintf("Session ID: %s", result.SessionID))
		output.Info(fmt.Sprintf("Request ID: %s", result.RequestID))
	}

	return nil
}

