# Validate that the rendered configuration.nix is parseable Nix.
# This test evaluates to true if the config can be imported without errors.
{ configPath ? "/tmp/test-config" }:

let
  config = import (configPath + "/configuration.nix");
in
  builtins.isFunction config
