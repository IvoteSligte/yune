{
  description = "Yune development flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShell = pkgs.mkShell {
          name = "antlr-shell";

          buildInputs = [
            pkgs.antlr4
          ];

          # Optional: convenience
          shellHook = ''
            echo "ANTLR version: $(antlr4 -version)"
          '';
        };
      }
    );
}
