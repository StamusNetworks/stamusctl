# nix/ — Nix Build Definitions

Nix expressions used by the flake to build stamusctl artifacts.

## iso.nix

Defines the **stamusctl live ISO** — a bootable NixOS image with a graphical desktop and stamusctl pre-installed. Built via:

```bash
nix build .#packages.x86_64-linux.iso
# or
make nix-iso
```

The ISO is based on `installation-cd-graphical-base.nix` from nixpkgs and includes:

-   **LXQt** desktop with **LightDM** (greeterless auto-login as `nixos`)
-   **stamusctl** available system-wide
-   Firewall disabled for easy network setup
-   Hostname: `stamusctl-live`
-   NixOS 25.05

The resulting ISO is written to `result/iso/stamusctl-live.iso`.

### Running the ISO

```bash
make nix-iso-run              # Build + launch in QEMU
stamusctl nix iso-run          # If already built
```

## Flake Overview

The root `flake.nix` uses this file and defines:

| Output                           | System       | Description                                             |
| -------------------------------- | ------------ | ------------------------------------------------------- |
| `packages.default`               | all          | stamusctl Go binary (static, CGO_ENABLED=0)             |
| `devShells.default`              | all          | Dev shell with go, golangci-lint, gofumpt, air, etc.    |
| `packages.x86_64-linux.iso`      | x86_64-linux | Live NixOS ISO (imports `nix/iso.nix`)                  |
| `checks.x86_64-linux.nixos-test` | x86_64-linux | VM integration test (imports `tests/nixos/vm-test.nix`) |
| `checks.x86_64-linux.iso-test`   | x86_64-linux | ISO graphical test (imports `tests/nixos/iso-test.nix`) |

Inputs: `nixpkgs` (nixos-25.05), `flake-utils`.
