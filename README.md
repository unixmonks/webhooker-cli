# webhooker-cli

A CLI tool for receiving webhooks locally during development. Connect to [webhooker.site](https://webhooker.site) and forward incoming webhooks to your local server.

## Installation

### From source

```bash
go install github.com/webhooker/webhooker-cli@latest
```

### From binary

Download the latest release from the [releases page](https://github.com/webhooker/webhooker-cli/releases).

### Build from source

```bash
git clone https://github.com/webhooker/webhooker-cli.git
cd webhooker-cli
make build
```

## Usage

```bash
webhooker connect <account-token> --forward <local-url>
```

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `--server` | Webhooker server URL | `wss://webhooker.site` |
| `--forward` | Local URL to forward webhooks to | (required) |
| `--verbose` | Enable verbose output | `false` |

### Examples

Forward webhooks to a local server on port 3000:

```bash
webhooker connect abc123 --forward http://localhost:3000
```

Forward to a specific endpoint with verbose output:

```bash
webhooker connect abc123 --forward http://localhost:8080/webhooks --verbose
```

Use a custom server:

```bash
webhooker connect abc123 --server wss://my-webhooker.example.com --forward http://localhost:3000
```

## How it works

1. Sign up at [webhooker.site](https://webhooker.site) and get your account token
2. Configure your external service to send webhooks to your unique webhook URL
3. Run the CLI to connect and forward webhooks to your local development server
4. Webhooks are forwarded in real-time via WebSocket connection

The CLI automatically reconnects if the connection is lost.

## License

MIT License - see [LICENSE](LICENSE) for details.
