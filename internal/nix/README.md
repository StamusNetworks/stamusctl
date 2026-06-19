# internal/nix — Nix Command Wrappers

Low-level Go wrappers around `nix`, `nixos-rebuild`, `nix-build`, `nix-instantiate`, and `qemu-system-x86_64`. This package is used by the handlers in `internal/handlers/nix/` and never by CLI commands directly.

## Functions

| Function                                    | Wraps                                | Description                                                                                                                      |
| ------------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| `IsNixOS()`                                 | checks `/etc/NIXOS`                  | Returns `true` if running on NixOS                                                                                               |
| `NixosRebuild(configPath, action)`          | `nixos-rebuild <action>`             | Runs nixos-rebuild with `-I nixos-config=<path>/configuration.nix`. Valid actions: `switch`, `boot`, `test`, `build`, `build-vm` |
| `BuildISO(configPath, outputDir)`           | `nix-build <nixpkgs/nixos>`          | Builds a NixOS ISO from `<configPath>/iso.nix`, outputs to `<outputDir>/result`                                                  |
| `RunShellTest(scriptPath, configPath)`      | `bash <script>`                      | Runs a shell test with `STAMUSCTL_CONFIG_PATH` set                                                                               |
| `RunNixTest(scriptPath, configPath)`        | `nix-instantiate --eval`             | Evaluates a Nix test expression, passing `configPath` as `--arg`                                                                 |
| `RunISO(isoPath, memory, cores, enableKVM)` | `qemu-system-x86_64`                 | Launches QEMU booting from the given ISO                                                                                         |
| `DiffClosures(currentSystem, newSystem)`    | `nix store diff-closures`            | Shows package-level diff between two store paths                                                                                 |
| `ListGenerations()`                         | `nixos-rebuild list-generations`     | Returns generation list as a string                                                                                              |
| `FindAndRunVM(resultDir)`                   | globs `run-*-vm` in `resultDir/bin/` | Finds and executes the VM launch script from a `nixos-rebuild build-vm` result                                                   |
| `Infect(configPath)`                        | —                                    | Stub, returns "not yet implemented"                                                                                              |

## Testing

The `execCommand` variable (defaults to `exec.Command`) can be replaced in tests to mock command execution. This avoids running real `nixos-rebuild` or `nix-build` during unit tests.

The filesystem is accessed via `app.FS` (afero), which can be swapped to an in-memory FS for testing `IsNixOS()` and other file-based checks.

See `nix_test.go` for examples.
