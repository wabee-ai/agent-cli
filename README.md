# Wabee CLI

A command-line interface for interacting with Wabee AI agents.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install wabee-ai/tap/wabee
```

### Go Install

```bash
go install github.com/wabee-ai/wabee-cli@latest
```

### Binary Download

Download pre-built binaries from the [releases page](https://github.com/wabee-ai/wabee-cli/releases).

### From Source

```bash
git clone https://github.com/wabee-ai/wabee-cli.git
cd wabee-cli
make install
```

## Quick Start

```bash
# Initialize configuration
wabee config init

# Check connection
wabee agent health

# Chat with the agent
wabee chat "What can you help me with?"

# View session history
wabee sessions list

# Debug a session
wabee trace <session-id>
```

## Usage

### Chat

```bash
# Simple message
wabee chat "What is the weather in NYC?"

# Continue a conversation
wabee chat --session abc123 "And tomorrow?"

# Stream response (default)
wabee chat --stream "Explain quantum computing"

# Pipe input
cat prompt.txt | wabee chat

# Output as JSON
wabee chat --output json "List 5 items"

# Interactive mode
wabee chat -i
```

### Sessions

```bash
# List sessions
wabee sessions list
wabee sessions list --limit 20

# Get session details
wabee sessions get <session-id>

# View execution trace
wabee sessions trace <session-id>

# Export session
wabee sessions export <session-id> --format json > session.json
wabee sessions export <session-id> --format markdown > session.md

# Delete session
wabee sessions delete <session-id>
```

### Execution Tracing

```bash
# View trace
wabee trace <session-id>

# Filter by request
wabee trace <session-id> --request req-001

# Show only tool calls
wabee trace <session-id> --tools-only

# Output formats
wabee trace <session-id> --output json
wabee trace <session-id> --output tree
```

### Configuration

```bash
# Initialize config
wabee config init

# Set values
wabee config set endpoint https://api.wabee.ai
wabee config set api-key <your-key>

# Show current config
wabee config show

# Manage profiles
wabee config profile list
wabee config profile create production
wabee config use production
```

### Agent Information

```bash
# Health check
wabee agent health

# Get metadata
wabee agent info

# List tools
wabee agent tools
```

## Configuration

Configuration is stored in `~/.wabee/config.yaml`:

```yaml
default_profile: development

profiles:
  development:
    endpoint: http://localhost:8000
    api_key: dev-key
    timeout: 30

  production:
    endpoint: https://api.wabee.ai
    api_key: ${WABEE_API_KEY}
    timeout: 60

defaults:
  output: text
  stream: true
  color: auto
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `WABEE_API_KEY` | API key (overrides config) |
| `WABEE_ENDPOINT` | API endpoint (overrides config) |
| `WABEE_PROFILE` | Default profile to use |
| `WABEE_CONFIG` | Path to config file |
| `NO_COLOR` | Disable colored output |

## Global Options

| Option | Short | Description |
|--------|-------|-------------|
| `--endpoint` | `-e` | API endpoint URL |
| `--api-key` | `-k` | API key |
| `--profile` | `-p` | Configuration profile |
| `--output` | `-o` | Output format (json, table, text, tree) |
| `--quiet` | `-q` | Suppress non-essential output |
| `--verbose` | `-v` | Verbose output |
| `--no-color` | | Disable colors |
| `--timeout` | | Request timeout (seconds) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Authentication error |
| 4 | Network error |
| 5 | Agent execution error |
| 6 | Invalid input |

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install Wabee CLI
  run: |
    curl -L https://github.com/wabee-ai/wabee-cli/releases/latest/download/wabee_linux_amd64.tar.gz | tar xz
    sudo mv wabee /usr/local/bin/

- name: Run Agent
  env:
    WABEE_API_KEY: ${{ secrets.WABEE_API_KEY }}
    WABEE_ENDPOINT: ${{ vars.WABEE_ENDPOINT }}
  run: wabee chat --output json "Analyze this PR"
```

## Development

```bash
# Build
make build

# Run tests
make test

# Lint
make lint

# Build for all platforms
make build-all
```

## License

MIT License - see [LICENSE](LICENSE) for details.
