{
  description = "Flake for stamusctl";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }: flake-utils.lib.eachDefaultSystem (system:
    let
      pkgs = import nixpkgs {
        inherit system;
      };

      stamusctl = pkgs.buildGoModule {
        pname = "stamusctl";
        version = "unstable";

        src = ./.;

        vendorHash = "sha256-RGgh+M1SM5wSZ0xHxnOuR/PydH6vPNH61BfAtyMFh1M=";

        subPackages = "cmd";
        CGO_ENABLED = 0;

        ldflags = [
          "-X stamus-ctl/internal/app.Arch=${system}"
          "-X stamus-ctl/internal/app.Commit=dev"
          "-X stamus-ctl/internal/app.Version=0.7.3"
          "-X stamus-ctl/internal/logging.envType=prd"
          "-extldflags=-static"
        ];

        postInstall = ''
          mv $out/bin/cmd $out/bin/stamusctl
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
}
