# internal/handlers/nix — NixOS Handler Layer

Business logic for all `stamusctl nix` subcommands. Each handler validates inputs, checks system state, and delegates to `internal/nix/` for command execution.

## Handlers

| Handler             | Called by      | Description                                                                                                                                            |
| ------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `NixInitHandler`    | `nix init`     | Pulls templates from registry (or embedded), renders configuration via the template engine, creates `values.yaml`, binds files, registers the instance |
| `NixSwitchHandler`  | `nix switch`   | Checks NixOS + config existence, creates automatic backup, calls `nixos-rebuild switch`                                                                |
| `NixTestHandler`    | `nix test`     | Discovers `.sh`/`.nix` test files in `<config>/tests/`, runs them, prints pass/fail summary                                                            |
| `NixISOHandler`     | `nix iso`      | Validates config, sanitizes output path, calls `nix-build` to produce an ISO                                                                           |
| `NixISORunHandler`  | `nix iso-run`  | Locates `*.iso` under `<output>/result/iso/`, launches in QEMU                                                                                         |
| `NixDiffHandler`    | `nix diff`     | Builds pending config, diffs against `/run/current-system`                                                                                             |
| `NixUpdateHandler`  | `nix update`   | Creates backup, pulls updated template, re-renders configuration                                                                                       |
| `NixBuildVMHandler` | `nix build-vm` | Calls `nixos-rebuild build-vm`, optionally launches the VM                                                                                             |
| `NixStatusHandler`  | `nix status`   | Prints config path/version/project, lists NixOS generations                                                                                            |
| `NixInfectHandler`  | `nix infect`   | Checks system is NOT NixOS, delegates to `nix.Infect()` (stub)                                                                                         |

## Common Patterns

- **NixOS gate**: handlers that require NixOS (`switch`, `test`, `diff`, `build-vm`) call `nix.IsNixOS()` first and return a clear error if not on NixOS.
- **Auto-backup**: destructive operations (`switch`, `update`) create a backup before proceeding.
- **Input structs**: each handler takes a typed `*HandlerInputs` struct, keeping the interface explicit and testable.
- **Exec mocking**: `NixUpdateHandler` uses a replaceable `execCommand` variable for testability, same pattern as `internal/nix/`.

## Tests

Handler tests live alongside the handlers (`*_test.go`). They use `internal/testutil` to swap `app.FS` to an in-memory filesystem and mock `nix.IsNixOS()` / exec calls to avoid requiring a real NixOS system.
