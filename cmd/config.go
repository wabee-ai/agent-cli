package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wabee-ai/wabee-cli/internal/config"
	"github.com/wabee-ai/wabee-cli/internal/output"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
	Long: `Manage CLI configuration including profiles and settings.

Configuration is stored in ~/.wabee/config.yaml by default.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long: `Initialize the CLI configuration interactively.

This will guide you through setting up your API endpoint and credentials.`,
	RunE: runConfigInit,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  endpoint    - API endpoint URL
  api-key     - API key for authentication
  auth-token  - Authentication token
  timeout     - Request timeout in seconds
  output      - Default output format (json, table, text, tree)
  stream      - Default streaming mode (true/false)
  color       - Color output mode (auto, always, never)

Examples:
  wabee config set endpoint https://api.wabee.ai
  wabee config set api-key your-api-key
  wabee config set timeout 120`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration settings.

Use --verbose to see all profiles and sensitive information.`,
	RunE: runConfigShow,
}

var configUseCmd = &cobra.Command{
	Use:   "use <profile>",
	Short: "Switch to a different profile",
	Long: `Switch to a different configuration profile.

Examples:
  wabee config use production
  wabee config use development`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigUse,
}

var configProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
}

var configProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available profiles",
	RunE:  runConfigProfileList,
}

var configProfileCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigProfileCreate,
}

var configProfileSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set the active profile",
	Long: `Set the active profile and persist the change.

This changes the default profile used for all subsequent commands.

Examples:
  wabee config profile set production
  wabee config profile set development`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigProfileSet,
}

var configProfileDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a profile",
	Long: `Delete a configuration profile.

You cannot delete the currently active profile. Switch to a different
profile first using 'wabee config profile set'.

Examples:
  wabee config profile delete old-profile
  wabee config profile delete staging`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigProfileDelete,
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configProfileCmd)

	configProfileCmd.AddCommand(configProfileListCmd)
	configProfileCmd.AddCommand(configProfileCreateCmd)
	configProfileCmd.AddCommand(configProfileSetCmd)
	configProfileCmd.AddCommand(configProfileDeleteCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to Wabee CLI!")
	fmt.Println()

	// Check if config already exists
	if config.Exists() {
		fmt.Print("Configuration already exists. Overwrite? [y/N]: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Get endpoint
	fmt.Print("Enter your API endpoint [http://localhost:8000]: ")
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	// Get API key
	fmt.Print("Enter API key (optional, for security we recommend using WABEE_API_KEY env var): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	// Create default profile
	profile := config.Profile{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Timeout:  60,
	}

	if err := config.CreateProfile("default", profile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	output.Success(fmt.Sprintf("Configuration saved to %s", config.ConfigFile()))
	fmt.Println()
	fmt.Println("You can now use the CLI:")
	fmt.Println("  wabee chat \"Hello, what can you help me with?\"")
	fmt.Println("  wabee agent health")

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	if err := config.SetValue(key, value); err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	output.Success(fmt.Sprintf("Set %s = %s", key, maskSensitive(key, value)))
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	profile := config.GetActiveProfile()

	fmt.Printf("Config file: %s\n", config.ConfigFile())
	fmt.Printf("Default profile: %s\n", cfg.DefaultProfile)
	fmt.Println()

	fmt.Println("Active profile settings:")
	fmt.Printf("  Endpoint: %s\n", profile.Endpoint)
	if profile.APIKey != "" {
		if isVerbose() {
			fmt.Printf("  API Key: %s\n", profile.APIKey)
		} else {
			fmt.Printf("  API Key: %s\n", maskString(profile.APIKey))
		}
	}
	if profile.AuthToken != "" {
		if isVerbose() {
			fmt.Printf("  Auth Token: %s\n", profile.AuthToken)
		} else {
			fmt.Printf("  Auth Token: %s\n", maskString(profile.AuthToken))
		}
	}
	fmt.Printf("  Timeout: %ds\n", profile.Timeout)
	fmt.Println()

	fmt.Println("Defaults:")
	fmt.Printf("  Output: %s\n", cfg.Defaults.Output)
	fmt.Printf("  Stream: %v\n", cfg.Defaults.Stream)
	fmt.Printf("  Color: %s\n", cfg.Defaults.Color)

	if isVerbose() && len(cfg.Profiles) > 1 {
		fmt.Println()
		fmt.Println("All profiles:")
		for name := range cfg.Profiles {
			marker := "  "
			if name == cfg.DefaultProfile {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, name)
		}
	}

	return nil
}

func runConfigUse(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	if err := config.UseProfile(profileName); err != nil {
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	output.Success(fmt.Sprintf("Switched to profile: %s", profileName))
	return nil
}

func runConfigProfileList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	profiles := config.ListProfiles()

	if len(profiles) == 0 {
		fmt.Println("No profiles configured.")
		fmt.Println("Run 'wabee config init' to create one.")
		return nil
	}

	for _, name := range profiles {
		marker := "  "
		if name == cfg.DefaultProfile {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, name)
	}

	return nil
}

func runConfigProfileCreate(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	name := args[0]

	fmt.Printf("Creating profile: %s\n", name)
	fmt.Println()

	// Get endpoint
	fmt.Print("Enter API endpoint: ")
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)

	// Get API key
	fmt.Print("Enter API key (optional, for security we recommend using WABEE_API_KEY env var): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	profile := config.Profile{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Timeout:  60,
	}

	if err := config.CreateProfile(name, profile); err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	output.Success(fmt.Sprintf("Profile '%s' created", name))

	// Ask if user wants to set the new profile as active
	fmt.Print("Set this profile as active? [y/N]: ")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		if err := config.UseProfile(name); err != nil {
			return fmt.Errorf("failed to set profile: %w", err)
		}
		output.Success(fmt.Sprintf("Active profile set to: %s", name))
	}

	return nil
}

func runConfigProfileSet(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	if err := config.UseProfile(profileName); err != nil {
		return fmt.Errorf("failed to set profile: %w", err)
	}

	output.Success(fmt.Sprintf("Active profile set to: %s", profileName))
	return nil
}

func runConfigProfileDelete(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	name := args[0]

	// Check if profile exists
	cfg := config.Get()
	if _, exists := cfg.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	// Prevent deleting the active profile
	if name == cfg.DefaultProfile {
		return fmt.Errorf("cannot delete the active profile '%s'. Switch to a different profile first using 'wabee config profile set'", name)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete profile '%s'? [y/N]: ", name)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	if err := config.DeleteProfile(name); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	output.Success(fmt.Sprintf("Profile '%s' deleted", name))
	return nil
}

func maskSensitive(key, value string) string {
	sensitiveKeys := map[string]bool{
		"api-key":    true,
		"api_key":    true,
		"auth-token": true,
		"auth_token": true,
		"token":      true,
		"secret":     true,
		"password":   true,
	}

	if sensitiveKeys[strings.ToLower(key)] {
		return maskString(value)
	}
	return value
}

func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
