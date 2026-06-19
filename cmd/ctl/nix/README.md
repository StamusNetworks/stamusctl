# stamusctl nix — CLI Commands

The `stamusctl nix` command family manages NixOS-based Stamus appliance deployments. It generates NixOS configuration from templates and applies it via `nixos-rebuild`.

## Subcommands

### `nix init`

Initialize a NixOS configuration by downloading templates and rendering configuration files.

```bash
stamusctl nix init                              # Default ClearNDR template
stamusctl nix init clearndr                     # Explicit ClearNDR template
stamusctl nix init suricata.interfaces=eth0     # With parameters
stamusctl nix init --version 1.2.3              # Specific template version
stamusctl nix init -v my-values.yaml            # From values file
```

| Flag         | Short | Default  | Description                                   |
| ------------ | ----- | -------- | --------------------------------------------- |
| `--default`  | `-d`  | `true`   | Use defaults (deprecated, now always true)    |
| `--expert`   | `-E`  | `false`  | Expert mode for advanced configuration        |
| `--values`   | `-v`  |          | Path to a values.yaml file                    |
| `--fromFile` | `-F`  |          | Use file content as parameter values          |
| `--config`   | `-c`  | `config` | Configuration output directory                |
| `--template` | `-t`  |          | Template folder path (hidden)                 |
| `--bind`     | `-b`  |          | Bind local files to config (`/local:/config`) |
| `--version`  |       | `latest` | Template version                              |
| `--registry` |       |          | Registry to pull templates from               |

Positional arguments: the first non-`key=value` argument selects the template to fetch from the registry (defaults to `clearndr`). Remaining `key=value` arguments set configuration parameters.

### `nix switch`

Apply the NixOS configuration via `nixos-rebuild switch`. Creates an automatic backup before switching.

```bash
stamusctl nix switch
stamusctl nix switch --config /path/to/config
```

Requires running on NixOS (checks `/etc/NIXOS`).

### `nix test`

Discover and run test scripts from the configuration's `tests/` directory.

-   `.sh` files are run with `bash` (receives `STAMUSCTL_CONFIG_PATH` env var)
-   `.nix` files are evaluated with `nix-instantiate --eval` (receives `configPath` arg)

```bash
stamusctl nix test
stamusctl nix test --config /path/to/config
stamusctl nix test --filter "check-*.sh"
```

| Flag       | Short | Default  | Description                       |
| ---------- | ----- | -------- | --------------------------------- |
| `--config` | `-c`  | `config` | Configuration directory           |
| `--filter` | `-f`  |          | Glob pattern to filter test files |

Exit code is non-zero if any test fails. Prints a pass/fail summary.

### `nix diff`

Preview package-level changes before switching. Builds the pending configuration and runs `nix store diff-closures` against `/run/current-system`.

```bash
stamusctl nix diff
stamusctl nix diff --config /path/to/config
```

Requires running on NixOS.

### `nix update`

Update NixOS configuration templates to a newer version. Creates an automatic backup before updating.

```bash
stamusctl nix update --version 1.3.0
stamusctl nix update --config /path/to/config
```

| Flag            | Short | Default  | Description                   |
| --------------- | ----- | -------- | ----------------------------- |
| `--config`      | `-c`  | `config` | Configuration directory       |
| `--version`     |       | `latest` | Target template version       |
| `--template`    | `-t`  |          | Template folder path (hidden) |
| `--interactive` |       | `false`  | Interactive parameter review  |

### `nix build-vm`

Build a throwaway QEMU VM from the NixOS configuration using `nixos-rebuild build-vm`.

```bash
stamusctl nix build-vm
stamusctl nix build-vm --run     # Build and launch immediately
```

| Flag       | Short | Default  | Description                  |
| ---------- | ----- | -------- | ---------------------------- |
| `--config` | `-c`  | `config` | Configuration directory      |
| `--run`    |       | `false`  | Launch the VM after building |

Requires running on NixOS.

### `nix status`

Display NixOS system and configuration status: config path, version, project name, and system generations.

```bash
stamusctl nix status
stamusctl nix status --config /path/to/config
```

### `nix iso`

Generate a NixOS ISO image from the configuration's `iso.nix` file using `nix-build`.

```bash
stamusctl nix iso
stamusctl nix iso --output /tmp/build
```

| Flag       | Short | Default  | Description                                 |
| ---------- | ----- | -------- | ------------------------------------------- |
| `--config` | `-c`  | `config` | Configuration directory                     |
| `--output` | `-o`  | `.`      | Output directory for the ISO result symlink |

### `nix iso-run`

Launch a previously built ISO in QEMU.

```bash
stamusctl nix iso-run
stamusctl nix iso-run --memory 8192 --cores 4
stamusctl nix iso-run --kvm=false
```

| Flag       | Short | Default  | Description                               |
| ---------- | ----- | -------- | ----------------------------------------- |
| `--config` | `-c`  | `config` | Configuration directory                   |
| `--output` | `-o`  | `.`      | Directory containing the ISO build result |
| `--memory` | `-m`  | `4096`   | RAM in megabytes                          |
| `--cores`  |       | `2`      | Number of CPU cores                       |
| `--kvm`    |       | `true`   | Enable KVM hardware acceleration          |

### `nix infect`

Convert an existing Linux system to NixOS. **Not yet implemented.**

## Code Structure

Each subcommand follows the project's command-handler separation:

-   **Command files** (this directory) define Cobra commands, register flags, and delegate to handlers.
-   **Handlers** (`internal/handlers/nix/`) contain the business logic.
-   **Nix package** (`internal/nix/`) wraps low-level nix/QEMU commands.
