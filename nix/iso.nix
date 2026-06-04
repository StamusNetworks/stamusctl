{ nixpkgs, stamusctl }:

let
  system = "x86_64-linux";

  nixosConfig = nixpkgs.lib.nixosSystem {
    inherit system;

    modules = [
      "${nixpkgs}/nixos/modules/installer/cd-dvd/installation-cd-graphical-base.nix"

      ({ pkgs, lib, ... }: {
        isoImage.edition = "stamusctl";
        isoImage.isoName = lib.mkForce "stamusctl-live.iso";

        # LXQt desktop environment
        services.xserver.desktopManager.lxqt.enable = true;

        # LightDM with greeterless auto-login
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

        # stamusctl available system-wide
        environment.systemPackages = [ stamusctl ];

        networking.hostName = "stamusctl-live";
        networking.firewall.enable = false;

        i18n.defaultLocale = "en_US.UTF-8";
        time.timeZone = "UTC";

        system.stateVersion = "25.05";
      })
    ];
  };
in
nixosConfig.config.system.build.isoImage
