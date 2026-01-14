# Wabee CLI

A command-line interface for interacting with Wabee AI agents.

## Installation

### Binary Download

Download pre-built binaries from the [releases page](https://github.com/wabee-ai/agent-cli/releases).

### From Source

```bash
git clone https://github.com/wabee-ai/agent-cli.git
cd agent-cli
make install
```

## Quick Start

```bash
# Initialize configuration
wabee config init

# Start a task with the agent
wabee task new "What can you help me with?"

# View session history
wabee sessions list

# Debug a session
wabee sessions trace <session-id>
```

## Usage

### Task

Start new tasks or continue conversations with the agent.

```bash
# Simple message
wabee task new "What is the weather in NYC?"

# Continue a conversation
wabee task followup --session abc123 "And tomorrow?"

# Stream response (enabled by default)
wabee task new --stream "Explain quantum computing"

# Disable streaming
wabee task new --stream=false "Quick question"

# Pipe input
cat prompt.txt | wabee task new

# Output as JSON
wabee task new --output json "List 5 items"
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
```

### Execution Tracing

View detailed execution traces for debugging agent behavior:

```bash
# View trace
wabee sessions trace <session-id>

# Filter by request
wabee sessions trace <session-id> --request-id req-001

# Show full content (no truncation)
wabee sessions trace <session-id> --full

# Set max content length (default 200)
wabee sessions trace <session-id> --max-content-length 500

# Output formats
wabee sessions trace <session-id> --output json
wabee sessions trace <session-id> --output tree
```

### Configuration

```bash
# Initialize config
wabee config init

# Set values
wabee config set endpoint https://api.wabee.ai
wabee config set api-key <your-key>
wabee config set auth-token <your-token>
wabee config set timeout 60
wabee config set output json
wabee config set stream true
wabee config set color auto

# Show current config
wabee config show

# Manage profiles
wabee config profile list
wabee config profile create production
wabee config profile set production
wabee config profile delete staging
wabee config use production
```

### Agent Information

```bash
# Get metadata
wabee agent info

# List tools
wabee agent tools
```

### Version

```bash
# Show version
wabee version

# Verbose version info (includes commit and build date)
wabee version -v
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
    auth_token: ${WABEE_AUTH_TOKEN}
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
| `--config` | | Path to config file (default: `~/.wabee/config.yaml`) |
| `--endpoint` | `-e` | API endpoint URL |
| `--api-key` | `-k` | API key |
| `--profile` | `-p` | Configuration profile |
| `--output` | `-o` | Output format (json, table, text, tree) |
| `--quiet` | `-q` | Suppress non-essential output |
| `--verbose` | `-v` | Verbose output |
| `--no-color` | | Disable colors |
| `--timeout` | | Request timeout (seconds) |

## CI/CD Integration

### GitHub Actions

```yaml
- name: Install Wabee CLI
  run: |
    curl -L https://github.com/wabee-ai/agent-cli/releases/latest/download/wabee_linux_amd64.tar.gz | tar xz
    sudo mv wabee /usr/local/bin/

- name: Run Agent
  env:
    WABEE_API_KEY: ${{ secrets.WABEE_API_KEY }}
    WABEE_ENDPOINT: ${{ vars.WABEE_ENDPOINT }}
  run: wabee task new --output json "Analyze this PR"
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
