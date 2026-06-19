# tests/nixos/ — NixOS Integration Tests

NixOS VM-based integration tests that validate stamusctl's nix commands and ISO builds. These run in ephemeral QEMU VMs via the NixOS test framework (`pkgs.testers.runNixOSTest`).

## Running

```bash
make nix-test-vm              # VM template test
make nix-test-iso             # ISO graphical test
make nix-test-syntax          # Validate .nix template syntax
make nix-test-vm-docker       # VM test inside a Docker container
make nix-test-cmd-docker      # Full init+test pipeline in Docker
```

Or directly via nix:

```bash
nix build .#checks.x86_64-linux.nixos-test -L
nix build .#checks.x86_64-linux.iso-test -L
```

Requires KVM (`/dev/kvm`). CI enables it via udev rules.

## Tests

### vm-test.nix — Template Rendering + nixos-rebuild

End-to-end test of the `stamusctl nix init` and `nix switch` pipeline:

1. Boots a NixOS VM with stamusctl and the test template pre-loaded
2. Runs `stamusctl nix init --config /tmp/test-config --default`
3. Verifies `configuration.nix` was rendered (no `{{ }}` markers remain)
4. Verifies the hostname parameter was injected (`stamusctl-test`)
5. Verifies `values.yaml` was created
6. Runs a real `nixos-rebuild build` to prove the rendered config is valid NixOS (store paths are pre-cached via `system.extraDependencies` so this is fast)
7. Runs `stamusctl nix switch` against a **stub** `nixos-rebuild` that validates argument passing (real switch would kill the test driver)

The test template lives in `fixtures/template/` and uses Go template syntax (`{{ .Values.hostname }}`), so this test exercises the full template rendering engine.

### iso-test.nix — ISO Graphical Environment

Validates the ISO's graphical environment works:

1. Boots a VM that mirrors the ISO configuration (LXQt + LightDM + auto-login)
2. Waits for the display manager and LXQt session to start
3. Verifies `stamusctl` is in `PATH`
4. Runs `stamusctl version` and `stamusctl --help`

## Fixtures

### fixtures/template/

Minimal template that exercises the Go template engine:

| File                             | Purpose                                                                   |
| -------------------------------- | ------------------------------------------------------------------------- |
| `config.yaml`                    | Declares a single `hostname` parameter (default: `stamusctl-test`)        |
| `configuration.nix`              | NixOS config with `{{ .Values.hostname }}` — must be rendered, not static |
| `tests/check-hostname.sh`        | Shell test: verifies `networking.hostName` is present in rendered config  |
| `tests/check-config-rendered.sh` | Shell test: verifies no `{{ }}` template markers remain                   |
| `tests/validate-syntax.nix`      | Nix test: imports rendered config and checks it's a valid Nix function    |

## CI

The GitHub Actions workflow (`.github/workflows/nix.yml`) runs these tests on every `workflow_call`:

1. **nix-template-syntax** — `nix-instantiate --parse` on all `.nix` files in `internal/embeds/`
2. **nix-vm-test** — `nix build .#checks.x86_64-linux.nixos-test`
3. **nix-iso-test** — `nix build .#checks.x86_64-linux.iso-test`
4. **nix-build-iso** — builds the final ISO (only after all tests pass), uploads as artifact (14-day retention)
