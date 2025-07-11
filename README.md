# Stamusctl

> **Important**: This repository contains the CLI tool only. Configuration templates are maintained in the separate [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates). Template-related issues should be reported there.

## Table of Contents

- [Description](#description)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Architecture](#architecture)
- [Template System](#template-system)
- [Commands Reference](#commands-reference)
- [Daemon Mode (stamusd)](#daemon-mode-stamusd)
- [Contributing](#contributing)
- [License](#license)
- [Support](#support)
- [Acknowledgments](#acknowledgments)

## Description

**stamusctl** is a comprehensive Command-Line Interface application written in Go by Stamus Networks that provides powerful functionalities to manage and deploy the Stamus security stack. It serves as the primary management tool for Stamus Network Detection and Response (NDR) deployments.

Key features include:

- **Configuration Management**: Manage Stamus stack configuration files with template-based approach
- **Stack Deployment**: Deploy and manage containerized Stamus stack using Docker Compose
- **Template System**: Built-in template system for configuration management with support for external templates
- **Registry Integration**: Support for public and private container registries
- **PCAP Analysis**: Built-in PCAP file reading capabilities for network traffic analysis
- **Daemon Mode**: REST API server (`stamusd`) for remote management

**stamusd** is the companion daemon that provides a REST API with functionalities similar to stamusctl, enabling remote management and automation.
You can find its API documentation [here](./cmd/daemon/docs/swagger.json).

For comprehensive documentation, visit [https://docs.clearndr.io/](https://docs.clearndr.io/).

## Installation

stamusctl can be installed through multiple methods:

### Quick Installation

**Direct Download:**

```bash
# Download the latest release
wget https://github.com/StamusNetworks/stamusctl/releases/latest/download/stamusctl-linux-amd64
chmod +x stamusctl-linux-amd64
sudo mv stamusctl-linux-amd64 /usr/local/bin/stamusctl
```

**From Stamus Networks:**

```bash
# Download from Stamus Networks
wget https://dl.clearndr.io/stamusctl-linux-amd64
chmod +x stamusctl-linux-amd64
sudo mv stamusctl-linux-amd64 /usr/local/bin/stamusctl
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/StamusNetworks/stamusctl.git
cd stamusctl

# Build stamusctl CLI
STAMUS_APP_NAME=stamusctl go build -o ./stamusctl ./cmd

# Build stamusd daemon
go build -o ./stamusd ./cmd
```

**Using Make:**

```bash
# Build both CLI and daemon
make all

# Build only CLI
make cli

# Build only daemon
make daemon
```

**Windows Build:**

```powershell
make all CLI_NAME=stamusctl.exe DAEMON_NAME=./stamusd.exe CURRENT_DIR=. HOST_OS=windows HOST_ARCH=amd64 VERSION=$(cat VERSION) GIT_COMMIT=$(git rev-parse HEAD)
```

### Nix/NixOS

For NixOS users, a Nix flake is available for integration into your system configuration:

```nix
# In your flake.nix inputs
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    stamusctl.url = "github:StamusNetworks/stamusctl";
    # ... other inputs
  };

  outputs = { self, nixpkgs, stamusctl, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        # ... your other modules
        {
          environment.systemPackages = with pkgs; [
            stamusctl.packages.x86_64-linux.default
            # ... other packages
          ];
        }
      ];
    };
  };
}
```

Or install directly for testing:

```bash
# Install using nix profile
nix profile install github:StamusNetworks/stamusctl

# Or run directly
nix run github:StamusNetworks/stamusctl
```

## Quick Start

### Basic Usage

```bash
# Initialize a new configuration with default settings
stamusctl compose init

# Initialize with custom network interface
stamusctl compose init suricata.interfaces=eth0

# Get current configuration
stamusctl config get

# Update configuration parameters
stamusctl config set suricata.interfaces=eth1

# Start the stack in detached mode
stamusctl compose up -d

# Process a PCAP file
stamusctl compose readpcap /path/to/capture.pcap

# Stop the stack
stamusctl compose down
```

### Advanced Configuration

```bash
# Initialize with custom values file
stamusctl compose init --values values.yaml

# Use custom configuration directory
stamusctl compose init --config /path/to/config

# Login to private registry
stamusctl login --registry registry.example.com --user username --pass password

# Update stack to specific version
stamusctl compose update --version v2.1.0

# Get template configuration keys
stamusctl template keys --template /path/to/template
```

## Architecture

stamusctl leverages a template-based architecture to manage complex security stack configurations:

- **Template Consumer**: stamusctl consumes templates from the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates)
- **Separation of Concerns**: Templates are maintained separately from the CLI tool for independent development
- **Registry Support**: Templates can be pulled from public/private container registries
- **Go Templates**: Uses Go's powerful template engine with Sprig functions for dynamic configuration

### Template Repository Structure

Templates are organized in the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates):

- Template definitions and configurations
- Version-specific templates
- Documentation for template usage
- **Issue tracking for template-related problems**

## Template System

stamusctl is a **template consumer** that uses configuration templates from the separate [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates). This separation allows for:

- **Independent Template Development**: Templates are maintained separately from the CLI tool
- **Template Versioning**: Templates can be updated independently of stamusctl releases
- **Template Issues**: **All template-related issues, bugs, and feature requests should be reported to the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates/issues), not this repository**

### How Templates Work

stamusctl consumes templates to generate configuration files:

- **External Templates**: Primary templates are fetched from the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates)
- **Default Template**: The default template used is "clearndr", hosted at [https://github.com/StamusNetworks/stamusctl-templates/tree/main/data/clearndr](https://github.com/StamusNetworks/stamusctl-templates/tree/main/data/clearndr)
- **Registry Integration**: Templates can be pulled from container registries
- **Version Management**: Support for different template versions via `--version` flag

## Usage

If you have the binary in your path, you can:

```
stamusctl [commands] [flags] [args]
```

If not, you can:

```
./stamusctl [commands] [flags] [args]
```

## Commands Reference

### Core Commands

#### `stamusctl compose`

Manages containerized Stamus stack deployments using Docker Compose.

**Subcommands:**

- `init` - Initialize configuration files

    - `--config` / `-c` - Configuration directory path (default: "config")
    - `--values` / `-v` - Use values.yaml file for configuration
    - `--fromFile` / `-F` - Use file content as parameter values
    - `--default` / `-d` - Use default settings (deprecated - now default behavior)
    - `--expert` / `-E` - Enable expert mode for advanced configuration
    - `--template` / `-t` - Specify template folder path (hidden)
    - `--version` - Target template version (default: "latest")
    - `--registry` - Registry to pull templates from
    - `--bind` / `-b` - Bind local files to config folder (`/local:/config`)
    - `[key]=[value]` - Set configuration values (e.g., `suricata.interfaces=eth0`)
    - `clearndr` - Initialize ClearNDR container compose file (subcommand)

- `update` - Update configuration to newer version

    - `--config` / `-c` - Configuration directory path
    - `--version` - Target version (default: "latest")
    - `--template` / `-t` - Template folder path (hidden)

- `readpcap` - Process PCAP files
    - `--config` / `-c` - Configuration directory path
    - `<pcap_file>` - Path to PCAP file (required argument)

**Docker Compose Commands:**

- `up` - Start the stack

    - `--config` / `-c` - Configuration directory path
    - `--detach` - Run in detached mode
    - `--build` - Build images before starting

- `down` - Stop the stack

    - `--config` / `-c` - Configuration directory path
    - `--volumes` - Remove named volumes
    - `--remove-orphans` - Remove undefined containers

- `restart` - Restart services

    - `--config` / `-c` - Configuration directory path

- `ps` - List running containers

    - `--config` / `-c` - Configuration directory path
    - `--services` - Display services
    - `--quiet` - Only display container IDs
    - `--format` - Format output

- `logs` - View service logs

    - `--config` / `-c` - Configuration directory path
    - `--timestamps` - Show timestamps
    - `--tail` - Number of lines to show from end
    - `--since` - Show logs since timestamp
    - `--until` - Show logs until timestamp
    - `--follow` - Follow log output
    - `--details` - Show extra details

- `exec` - Execute command in running container

    - `--config` / `-c` - Configuration directory path
    - `--detach` - Run in detached mode
    - `--privileged` - Give extended privileges
    - `--user` - Username or UID
    - `--workdir` - Working directory
    - `--env` - Set environment variables
    - `--no-TTY` - Disable pseudo-TTY allocation
    - `--dry-run` - Execute command in dry run mode
    - `--index` - Index of container if multiple instances

- `pull` - Pull service images

    - `--config` / `-c` - Configuration directory path
    - `--ignore-buildable` - Ignore buildable images
    - `--ignore-pull-failures` - Ignore pull failures
    - `--include-deps` - Include dependencies
    - `--quiet` - Pull without printing progress

- `images` - List images
    - `--config` / `-c` - Configuration directory path
    - `--format` - Format output
    - `--quiet` - Only display image IDs

#### `stamusctl config`

Manages configuration files and parameters.

**Subcommands:**

- `get` - Display configuration values (also default when no subcommand specified)

    - `--config` / `-c` - Configuration directory path
    - `[key]...` - Get specific configuration values
    - `content` - Get configuration file architecture
    - `keys` - Get configuration parameter keys
        - `--markdown` - Output in Markdown format

- `set` - Modify configuration

    - `--config` / `-c` - Configuration directory path
    - `--values` / `-v` - Values file to use
    - `--reload` - Reload configuration, don't keep arbitrary parameters
    - `--apply` - Apply configuration changes
    - `--fromFile` / `-F` - Use file content as parameter values
    - `[key]=[value]...` - Set configuration values
    - `content` - Set configuration files
        - `[host_folder]:[config_folder]` - Bind specific configuration files

- `list` - List available configurations

- `clear` - Clear configuration values

    - `--config` / `-c` - Configuration directory path

- `version` - Get configuration version
    - `--config` / `-c` - Configuration directory path

#### `stamusctl template`

Manages configuration templates.

**Subcommands:**

- `keys` - List available template keys
    - `--template` / `-t` - Template folder path (required)
    - `--markdown` - Output in Markdown format

#### `stamusctl login`

Manages authentication with container registries.

**Options:**

- `--registry` - Registry URL (required)
- `--user` - Username (required)
- `--pass` - Password (required)
- `--verif` - Verify registry connectivity (default: true)

#### `stamusctl version`

Display version information including build commit and architecture.

### Daemon Commands (`stamusd`)

#### `stamusd run`

Start the daemon REST API server.

**Options:**

- `--verbose` - Set verbosity level (0-3)

#### `stamusd version`

Display daemon version information.

### Global Options

- `--verbose` - Set verbosity level (0-3)
- `--help` - Show help information

## Daemon Mode (stamusd)

The `stamusd` daemon provides a REST API for remote management of Stamus stacks. It offers similar functionality to the CLI but through HTTP endpoints.

### Starting the Daemon

```bash
# Start the daemon
./stamusd run

# Start with custom configuration
./stamusd run --config /path/to/config
```

### API Documentation

The daemon provides a Swagger-documented REST API. Access the documentation at:

- Swagger JSON: `./cmd/daemon/docs/swagger.json`
- When running: `http://localhost:8080/swagger/index.html`

## Contributing

We welcome contributions to stamusctl! Here's how you can help:

### Getting Started

1. **Fork the Repository**

    ```bash
    git clone https://github.com/StamusNetworks/stamusctl.git
    cd stamusctl
    ```

2. **Set Up Development Environment**

    ```bash
    # Install dependencies
    go mod download

    # Install development tools
    go install github.com/cosmtrek/air@latest  # For live reloading
    go install github.com/swaggo/swag/cmd/swag@latest  # For API docs
    ```

3. **Build and Test**

    ```bash
    # Build the project
    make all

    # Run tests
    make test

    # Run with coverage
    make cover

    # Run linter
    golangci-lint run
    ```

### Development Guidelines

#### Code Style

- Follow Go best practices and conventions
- Use `gofmt` for code formatting
- Write comprehensive tests for new features
- Include documentation for public APIs

#### Template Development

- **Templates are maintained in a separate repository**: [stamusctl-templates](https://github.com/StamusNetworks/stamusctl-templates)
- **Template contributions should be made to that repository, not this one**
- Follow the template structure and naming conventions defined in the template repository
- Test templates with different configuration scenarios
- **Report template issues to the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates/issues)**

#### Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): description

[optional body]

[optional footer]
```

Examples:

- `feat: add template validation command`
- `fix: resolve configuration parsing issue`
- `docs: update installation instructions`

#### Testing

- Write unit tests for new functionality
- Test with external templates from the template repository
- Verify Docker Compose integration
- Test registry authentication

#### Pull Request Process

1. **Create a Feature Branch**

    ```bash
    git checkout -b feature/your-feature-name
    ```

2. **Make Your Changes**

    - Implement your feature or fix
    - Add/update tests as needed
    - Update documentation

3. **Test Your Changes**

    ```bash
    make test
    make cover
    ```

4. **Submit Pull Request**
    - Provide clear description of changes
    - Reference related issues
    - Include testing steps

### Areas for Contribution

- **CLI Features**: New commands, improved registry integration, authentication improvements
- **Bug Fixes**: Issues with configuration parsing, Docker integration, CLI functionality
- **Documentation**: Examples, use cases, API documentation, CLI help text
- **Testing**: Unit tests, integration tests, functional tests
- **Templates**: **Template-related contributions should be made to the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates)**

### Reporting Issues

When reporting bugs or requesting features:

1. **Template Issues**: Report template-related problems to the [stamusctl-templates repository](https://github.com/StamusNetworks/stamusctl-templates/issues)
2. **CLI Issues**: Report CLI tool issues to this repository
3. Check existing issues first
4. Use the appropriate issue template
5. Provide detailed reproduction steps
6. Include version information (`stamusctl version`)
7. Share relevant configuration files (sanitized)

### Issue Categories

- **stamusctl Repository**: CLI bugs, new CLI commands, authentication, registry integration, Docker Compose wrapper issues
- **stamusctl-templates Repository**: Template configuration issues, template bugs, new template requests, template documentation

### Development Environment

#### Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- Git

#### Optional Tools

- `air` for live reloading during development
- `golangci-lint` for code quality checks
- `make` for build automation

#### Project Structure

```
stamusctl/
├── cmd/                 # Command-line interfaces
│   ├── ctl/            # CLI commands
│   └── daemon/         # Daemon API
├── internal/           # Internal packages
│   ├── handlers/       # Command handlers
│   ├── models/         # Data models
│   └── docker/         # Docker integration
├── pkg/                # Public packages
└── scripts/            # Build and utility scripts
```

### Code of Conduct

Please be respectful and inclusive when participating in the project. Follow our code of conduct and help create a welcoming environment for all contributors.

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0). See the [LICENSE](LICENSE) file for details.

The GPL-3.0 license ensures that:

- You can use, modify, and distribute the software
- Any modifications must also be released under GPL-3.0
- The software is provided without warranty

For commercial licensing or questions about the license, please contact [Stamus Networks](https://www.stamus-networks.com/).

## Support

- **Documentation**:
    - Complete documentation at [https://docs.clearndr.io/](https://docs.clearndr.io/)
    - Inline help (`stamusctl --help`)
    - This README for quick reference
- **Issues**:
    - **CLI Tool Issues**: Report bugs and request features on [GitHub Issues](https://github.com/StamusNetworks/stamusctl/issues)
    - **Template Issues**: Report template problems to [stamusctl-templates Issues](https://github.com/StamusNetworks/stamusctl-templates/issues)
- **Community**: Join discussions in our community channels
- **Professional Support**: Contact [Stamus Networks](https://www.stamus-networks.com/) for enterprise support

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI functionality
- Uses [Gin](https://github.com/gin-gonic/gin) for REST API
- Template system powered by Go templates and [Sprig](https://github.com/Masterminds/sprig)
- Docker integration via [Docker Compose](https://github.com/docker/compose)
