{ pkgs, stamusctl }:

{
  name = "stamusctl-iso-graphical";

  nodes.machine = { config, pkgs, lib, ... }: {
    # Mirror the ISO configuration: LXQt + LightDM + stamusctl
    services.xserver.enable = true;
    services.xserver.desktopManager.lxqt.enable = true;

    services.xserver.displayManager.lightdm = {
      enable = true;
      greeter.enable = false;
      autoLogin.timeout = 0;
    };

    services.displayManager = {
      defaultSession = "lxqt";
      autoLogin = {
        enable = true;
        user = "nixos";
      };
    };

    users.users.nixos = {
      isNormalUser = true;
      extraGroups = [ "wheel" ];
      password = "";
    };

    environment.systemPackages = [ stamusctl ];

    networking.hostName = "stamusctl-live";
    networking.firewall.enable = false;

    virtualisation = {
      memorySize = 4096;
      diskSize = 4096;
      cores = 2;
    };

    system.stateVersion = "25.05";
  };

  testScript = ''
    machine.start()
    machine.wait_for_unit("multi-user.target")

    # 1. Verify display manager starts
    machine.wait_for_unit("display-manager.service")

    # 2. Wait for LXQt desktop session (auto-login)
    machine.wait_until_succeeds("pgrep -u nixos lxqt-session", timeout=60)

    # 3. Verify stamusctl is in PATH
    machine.succeed("which stamusctl")

    # 4. Verify stamusctl runs and outputs version
    output = machine.succeed("stamusctl version")
    print(f"stamusctl version: {output}")
    assert output.strip() != "", "stamusctl version returned empty output"

    # 5. Verify stamusctl help works (basic CLI sanity)
    machine.succeed("stamusctl --help")
  '';
}
