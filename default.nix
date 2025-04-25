{ pkgs ? import <nixpkgs> { } }:
pkgs.mkShell {
  buildInputs = with pkgs; [ go golangci-lint gotools gofumpt air go-swag hadolint trivy ];
}
