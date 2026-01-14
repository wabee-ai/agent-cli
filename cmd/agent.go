package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wabee-ai/wabee-cli/internal/client"
	"github.com/wabee-ai/wabee-cli/internal/output"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent information and management",
	Long:  `Commands for getting agent information and health status.`,
}

var agentInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get agent metadata",
	Long: `Display metadata about the connected agent.

Examples:
  wabee agent info
  wabee agent info --output json`,
	RunE: runAgentInfo,
}

var agentToolsCmd = &cobra.Command{
	Use:     "tools",
	Aliases: []string{"tool"},
	Short:   "List available tools",
	Long: `List all tools available to the agent.

Examples:
  wabee agent tools
  wabee agent tools --output json
  wabee agent tools --output table`,
	RunE: runAgentTools,
}

func init() {
	agentCmd.AddCommand(agentInfoCmd)
	agentCmd.AddCommand(agentToolsCmd)
}

func runAgentInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()
	formatter := getFormatter()

	info, err := apiClient.GetAgentInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get agent info: %w", err)
	}

	return formatter.Output(info)
}

func runAgentTools(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	apiClient := client.New()
	formatter := getFormatter()

	tools, err := apiClient.GetTools(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tools: %w", err)
	}

	if len(tools) == 0 {
		if !isQuiet() {
			output.Warn("No tools available")
		}
		return nil
	}

	return formatter.Output(tools)
}
