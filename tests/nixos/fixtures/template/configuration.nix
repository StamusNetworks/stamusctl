{ config, pkgs, ... }:
{
  networking.hostName = "{{ .Values.hostname }}";

  fileSystems."/" = { device = "/dev/vda1"; fsType = "ext4"; };
  boot.loader.grub.device = "nodev";
  system.stateVersion = "25.05";

  # Marker to verify the switch succeeded and templating worked
  environment.etc."stamusctl-test-marker".text = "switch-succeeded-{{ .Values.hostname }}";
}
