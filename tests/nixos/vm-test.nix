{ pkgs, stamusctl }:

let
  # Pre-build the target NixOS system on the host so all store paths are
  # available inside the VM. nixos-rebuild build will find everything cached.
  targetSystem = (pkgs.nixos {
    networking.hostName = "stamusctl-test";
    fileSystems."/" = { device = "/dev/vda1"; fsType = "ext4"; };
    boot.loader.grub.device = "nodev";
    system.stateVersion = "25.05";
    environment.etc."stamusctl-test-marker".text = "switch-succeeded-stamusctl-test";
  }).toplevel;

  # Stub nixos-rebuild for the switch step. We can't run real switch because
  # it replaces the running system and kills the test driver's communication.
  # The stub validates stamusctl passes the correct arguments.
  nixos-rebuild-stub = pkgs.writeShellScriptBin "nixos-rebuild" ''
    set -euo pipefail

    if [ "$1" != "switch" ]; then
      echo "ERROR: expected action 'switch', got '$1'" >&2
      exit 1
    fi

    CONFIG_PATH=""
    shift
    while [ $# -gt 0 ]; do
      case "$1" in
        -I)
          shift
          if [[ "$1" == nixos-config=* ]]; then
            CONFIG_PATH="''${1#nixos-config=}"
          fi
          ;;
      esac
      shift
    done

    if [ -z "$CONFIG_PATH" ]; then
      echo "ERROR: no -I nixos-config=<path> argument found" >&2
      exit 1
    fi

    if [ ! -f "$CONFIG_PATH" ]; then
      echo "ERROR: configuration file not found: $CONFIG_PATH" >&2
      exit 1
    fi

    echo "nixos-rebuild stub: config validated at $CONFIG_PATH"
  '';
in
{
  name = "stamusctl-nix-commands";

  nodes.machine = { config, pkgs, lib, ... }: {
    environment.systemPackages = [ stamusctl ];

    # nixos-rebuild build needs NIX_PATH to find <nixpkgs>
    nix.nixPath = [ "nixpkgs=${pkgs.path}" ];

    # Pre-populate the VM's store with the target system closure
    system.extraDependencies = [ targetSystem ];

    # Place the test template where the test script can copy it
    environment.etc."stamusctl-test-template/config.yaml".source =
      ./fixtures/template/config.yaml;
    environment.etc."stamusctl-test-template/configuration.nix".source =
      ./fixtures/template/configuration.nix;

    virtualisation = {
      memorySize = 2048;
      diskSize = 4096;
      cores = 2;
    };
  };

  testScript = let
    stubBin = "${nixos-rebuild-stub}/bin";
  in ''
    machine.start()
    machine.wait_for_unit("multi-user.target")

    # 1. Verify stamusctl is installed and runs
    machine.succeed("stamusctl version")

    # 2. Verify NixOS detection works
    machine.succeed("test -f /etc/NIXOS")

    # 3. Pre-place template at DefaultClearNDRPath
    machine.succeed(
        "mkdir -p /tmp/stamus-templates/clearndr/embedded/"
    )
    machine.succeed(
        "cp /etc/stamusctl-test-template/* /tmp/stamus-templates/clearndr/embedded/"
    )

    # 4. Run nix init — full template rendering pipeline
    machine.succeed(
        "EMBED_MODE=true "
        "STAMUS_TEMPLATES_FOLDER=/tmp/stamus-templates/ "
        "stamusctl nix init --config /tmp/test-config --default"
    )

    # 5. Verify template rendering produced configuration.nix with injected values
    output = machine.succeed("cat /tmp/test-config/configuration.nix")
    assert "stamusctl-test" in output, f"Template rendering failed: {output}"
    assert "{{" not in output, f"Unrendered template markers: {output}"

    # 6. Verify values.yaml was created
    machine.succeed("test -f /tmp/test-config/values.yaml")

    # 7. Prove the rendered config is valid NixOS — real nixos-rebuild build
    #    evaluates and builds the system closure. All store paths are pre-cached
    #    via system.extraDependencies so this is fast (~10s).
    #    We can't use "switch" because it replaces the running system and kills
    #    the test driver's communication channel.
    machine.succeed(
        "nixos-rebuild build "
        "-I nixos-config=/tmp/test-config/configuration.nix "
        "> /tmp/build.log 2>&1"
    )

    # 8. Test nix switch handler logic — stub validates stamusctl passes
    #    correct arguments (action=switch, -I nixos-config=<path>)
    machine.succeed(
        "PATH=${stubBin}:$PATH "
        "stamusctl nix switch --config /tmp/test-config"
    )
  '';
}
