package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the CLI configuration
type Config struct {
	DefaultProfile string             `mapstructure:"default_profile" yaml:"default_profile"`
	Profiles       map[string]Profile `mapstructure:"profiles" yaml:"profiles"`
	Defaults       Defaults           `mapstructure:"defaults" yaml:"defaults"`
}

// Profile represents a named configuration profile
type Profile struct {
	Endpoint  string `mapstructure:"endpoint" yaml:"endpoint"`
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`
	AuthToken string `mapstructure:"auth_token" yaml:"auth_token"`
	Timeout   int    `mapstructure:"timeout" yaml:"timeout"`
}

// Defaults represents default settings
type Defaults struct {
	Output string `mapstructure:"output" yaml:"output"`
	Stream bool   `mapstructure:"stream" yaml:"stream"`
	Color  string `mapstructure:"color" yaml:"color"`
}

var (
	cfg         *Config
	cfgFile     string
	profileName string
)

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".wabee"
	}
	return filepath.Join(home, ".wabee")
}

// ConfigFile returns the configuration file path
func ConfigFile() string {
	if cfgFile != "" {
		return cfgFile
	}
	return filepath.Join(ConfigDir(), "config.yaml")
}

// SetConfigFile sets a custom config file path
func SetConfigFile(path string) {
	cfgFile = path
}

// SetProfile sets the active profile name
func SetProfile(name string) {
	profileName = name
}

// Init initializes the configuration
func Init() error {
	viper.SetConfigType("yaml")

	// Set config file path
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(ConfigDir())
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	// Environment variable bindings
	viper.SetEnvPrefix("WABEE")
	viper.AutomaticEnv()

	// Bind specific environment variables
	viper.BindEnv("api_key", "WABEE_API_KEY")
	viper.BindEnv("endpoint", "WABEE_ENDPOINT")
	viper.BindEnv("profile", "WABEE_PROFILE")

	// Set defaults
	viper.SetDefault("default_profile", "default")
	viper.SetDefault("defaults.output", "text")
	viper.SetDefault("defaults.stream", true)
	viper.SetDefault("defaults.color", "auto")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	// Ensure profiles map exists
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return nil
}

// Get returns the current configuration
func Get() *Config {
	if cfg == nil {
		cfg = &Config{
			DefaultProfile: "default",
			Profiles:       make(map[string]Profile),
			Defaults: Defaults{
				Output: "text",
				Stream: true,
				Color:  "auto",
			},
		}
	}
	return cfg
}

// GetActiveProfile returns the currently active profile
func GetActiveProfile() Profile {
	config := Get()

	// Check for environment variable override
	if envProfile := viper.GetString("profile"); envProfile != "" {
		profileName = envProfile
	}

	// Determine which profile to use
	name := profileName
	if name == "" {
		name = config.DefaultProfile
	}
	if name == "" {
		name = "default"
	}

	profile, exists := config.Profiles[name]
	if !exists {
		profile = Profile{}
	}

	// Apply environment variable overrides
	if apiKey := viper.GetString("api_key"); apiKey != "" {
		profile.APIKey = apiKey
	}
	if endpoint := viper.GetString("endpoint"); endpoint != "" {
		profile.Endpoint = endpoint
	}

	// Set default timeout if not specified
	if profile.Timeout == 0 {
		profile.Timeout = 60
	}

	return profile
}

// GetEndpoint returns the API endpoint
func GetEndpoint() string {
	profile := GetActiveProfile()
	if profile.Endpoint == "" {
		return "http://localhost:8000"
	}
	return profile.Endpoint
}

// GetAPIKey returns the API key
func GetAPIKey() string {
	return GetActiveProfile().APIKey
}

// GetAuthToken returns the auth token
func GetAuthToken() string {
	return GetActiveProfile().AuthToken
}

// GetTimeout returns the request timeout in seconds
func GetTimeout() int {
	return GetActiveProfile().Timeout
}

// GetOutputFormat returns the default output format
func GetOutputFormat() string {
	config := Get()
	if config.Defaults.Output == "" {
		return "text"
	}
	return config.Defaults.Output
}

// GetStreamDefault returns the default streaming setting
func GetStreamDefault() bool {
	return Get().Defaults.Stream
}

// UseColor returns whether to use colored output
func UseColor() bool {
	config := Get()
	switch config.Defaults.Color {
	case "never", "false", "no":
		return false
	case "always", "true", "yes":
		return true
	default: // "auto"
		// Check NO_COLOR environment variable
		if os.Getenv("NO_COLOR") != "" {
			return false
		}
		// Check if stdout is a terminal
		fi, _ := os.Stdout.Stat()
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
}

// Save saves the current configuration to file
func Save() error {
	configDir := ConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("default_profile", cfg.DefaultProfile)
	viper.Set("profiles", cfg.Profiles)
	viper.Set("defaults", cfg.Defaults)

	configFile := ConfigFile()
	return viper.WriteConfigAs(configFile)
}

// SetValue sets a configuration value
func SetValue(key, value string) error {
	config := Get()
	profile := GetActiveProfile()
	profileName := config.DefaultProfile
	if profileName == "" {
		profileName = "default"
	}

	switch key {
	case "endpoint":
		profile.Endpoint = NormalizeEndpoint(value)
	case "api-key", "api_key":
		profile.APIKey = value
	case "auth-token", "auth_token":
		profile.AuthToken = value
	case "timeout":
		fmt.Sscanf(value, "%d", &profile.Timeout)
	case "output":
		config.Defaults.Output = value
	case "stream":
		config.Defaults.Stream = value == "true" || value == "yes" || value == "1"
	case "color":
		config.Defaults.Color = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	config.Profiles[profileName] = profile
	return Save()
}

// CreateProfile creates a new profile
func CreateProfile(name string, profile Profile) error {
	config := Get()
	profile.Endpoint = NormalizeEndpoint(profile.Endpoint)
	config.Profiles[name] = profile
	return Save()
}

// UseProfile sets the default profile
func UseProfile(name string) error {
	config := Get()
	if _, exists := config.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	config.DefaultProfile = name
	return Save()
}

// ListProfiles returns all profile names
func ListProfiles() []string {
	config := Get()
	names := make([]string, 0, len(config.Profiles))
	for name := range config.Profiles {
		names = append(names, name)
	}
	return names
}

// DeleteProfile deletes a profile by name
func DeleteProfile(name string) error {
	config := Get()
	if _, exists := config.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' does not exist", name)
	}
	delete(config.Profiles, name)
	return Save()
}

// Exists returns true if a config file exists
func Exists() bool {
	_, err := os.Stat(ConfigFile())
	return err == nil
}

// NormalizeEndpoint removes /core or /core/v1 suffixes from the endpoint URL.
// The API client adds the /core/v1 prefix automatically.
func NormalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSuffix(endpoint, "/")
	endpoint = strings.TrimSuffix(endpoint, "/core/v1")
	endpoint = strings.TrimSuffix(endpoint, "/core")
	return endpoint
}
