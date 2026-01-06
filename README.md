# webhooker-cli

A CLI tool for receiving webhooks locally during development. Connect to [webhooker.site](https://webhooker.site) and forward incoming webhooks to your local server.

## Installation

### macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -sL https://github.com/unixmonks/webhooker-cli/releases/latest/download/webhooker-cli_darwin_arm64.tar.gz | tar xz
sudo mv webhooker-cli /usr/local/bin/webhooker

# Intel
curl -sL https://github.com/unixmonks/webhooker-cli/releases/latest/download/webhooker-cli_darwin_amd64.tar.gz | tar xz
sudo mv webhooker-cli /usr/local/bin/webhooker
```

### Linux

```bash
# x86_64
curl -sL https://github.com/unixmonks/webhooker-cli/releases/latest/download/webhooker-cli_linux_amd64.tar.gz | tar xz
sudo mv webhooker-cli /usr/local/bin/webhooker

# ARM64
curl -sL https://github.com/unixmonks/webhooker-cli/releases/latest/download/webhooker-cli_linux_arm64.tar.gz | tar xz
sudo mv webhooker-cli /usr/local/bin/webhooker
```

### Windows

Download the latest `.zip` from the [releases page](https://github.com/unixmonks/webhooker-cli/releases) and add to your PATH.

### From source

```bash
go install github.com/unixmonks/webhooker-cli@latest
```

### Build from source

```bash
git clone https://github.com/unixmonks/webhooker-cli.git
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
