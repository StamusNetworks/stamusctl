{
  description = "Flake for stamusctl";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    let
      perSystem = flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };

          version = builtins.replaceStrings [ "\n" ] [ "" ] (builtins.readFile ./VERSION);
          gitCommit = self.shortRev or "dirty";

          stamusctl = pkgs.buildGoModule {
            pname = "stamusctl";
            inherit version;

            src = ./.;

            vendorHash = "sha256-7J9ZkZpuu26z0N1/sDdQ6zn8EUMdBrmLadZX9EdnHi8=";

            subPackages = "cmd";
            CGO_ENABLED = 0;

            nativeBuildInputs = [ pkgs.installShellFiles ];

            ldflags = [
              "-X stamus-ctl/internal/app.Arch=${system}"
              "-X stamus-ctl/internal/app.Commit=${gitCommit}"
              "-X stamus-ctl/internal/app.Version=${version}"
              "-X stamus-ctl/internal/logging.envType=prd"
              "-extldflags=-static"
            ];

            postInstall = ''
              mv $out/bin/cmd $out/bin/stamusctl

              # Install shell completions
              installShellCompletion --cmd stamusctl \
                --bash <($out/bin/stamusctl completion bash) \
                --fish <($out/bin/stamusctl completion fish) \
                --zsh <($out/bin/stamusctl completion zsh)
            '';

            meta = with pkgs.lib; {
              description = "CLI for managing Stamus Security Platform";
              homepage = "https://github.com/StamusNetworks/stamusctl";
              license = licenses.mit;
            };
          };
        in
        {
          packages.default = stamusctl;
          apps.default = flake-utils.lib.mkApp {
            drv = stamusctl;
          };
          devShells.default = pkgs.mkShell {
            packages = with pkgs; [ go golangci-lint gotools gofumpt air go-swag hadolint trivy ];
          };
        }
      );

      # NixOS VM integration tests (Linux only)
      linuxChecks = let
        system = "x86_64-linux";
        pkgs = import nixpkgs { inherit system; };
      in {
        checks.${system} = {
          nixos-test = pkgs.testers.runNixOSTest (
            import ./tests/nixos/vm-test.nix {
              inherit pkgs;
              stamusctl = perSystem.packages.${system}.default;
            }
          );

          iso-test = pkgs.testers.runNixOSTest (
            import ./tests/nixos/iso-test.nix {
              inherit pkgs;
              stamusctl = perSystem.packages.${system}.default;
            }
          );
        };

        packages.${system}.iso = import ./nix/iso.nix {
          inherit nixpkgs;
          stamusctl = perSystem.packages.${system}.default;
        };
      };
    in
    nixpkgs.lib.recursiveUpdate perSystem linuxChecks;
}
