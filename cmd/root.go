package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wabee-ai/wabee-cli/internal/config"
	"github.com/wabee-ai/wabee-cli/internal/output"
)

var (
	// Global flags
	cfgFile     string
	profileFlag string
	outputFlag  string
	quietFlag   bool
	verboseFlag bool
	noColorFlag bool
	endpointFlag string
	apiKeyFlag   string
	timeoutFlag  int

	// Version information (set at build time)
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "wabee",
	Short: "Wabee Agent CLI - Interact with Wabee AI agents",
	Long: `Wabee CLI is a command-line tool for interacting with Wabee AI agents.

It allows you to:
  - Send tasks to agents using natural language
  - Manage and debug sessions
  - View execution traces
  - Configure multiple profiles for different environments

Example:
  wabee task new "What is the weather in NYC?"
  wabee task followup --session <id> "And tomorrow?"
  wabee sessions list
  wabee sessions trace <session-id>`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Apply flag overrides
		if cfgFile != "" {
			config.SetConfigFile(cfgFile)
		}
		if profileFlag != "" {
			config.SetProfile(profileFlag)
		}

		// Initialize configuration
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		return nil
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wabee/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&profileFlag, "profile", "p", "", "configuration profile to use")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output format (json, table, text, tree)")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output for debugging")
	rootCmd.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().StringVarP(&endpointFlag, "endpoint", "e", "", "API endpoint URL")
	rootCmd.PersistentFlags().StringVarP(&apiKeyFlag, "api-key", "k", "", "API key for authentication")
	rootCmd.PersistentFlags().IntVar(&timeoutFlag, "timeout", 0, "request timeout in seconds")

	// Disable auto-generated completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add subcommands
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(sessionsCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(versionCmd)
}

// versionCmd shows version information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wabee version %s\n", Version)
		if verboseFlag {
			fmt.Printf("  Commit: %s\n", Commit)
			fmt.Printf("  Built: %s\n", BuildDate)
		}
	},
}

// getOutputFormat returns the output format to use
func getOutputFormat() string {
	if outputFlag != "" {
		return outputFlag
	}
	return config.GetOutputFormat()
}

// getFormatter returns a configured formatter
func getFormatter() *output.Formatter {
	return output.NewFormatter(getOutputFormat())
}

// isQuiet returns true if output should be suppressed
func isQuiet() bool {
	return quietFlag
}

// isVerbose returns true if verbose output is enabled
func isVerbose() bool {
	return verboseFlag
}
