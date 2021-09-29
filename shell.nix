{ nixpkgs ? <nixpkgs>, pkgs ? import nixpkgs { } }:
let pkgs = import nixpkgs { };
in pkgs.mkShell {
  name = "nix-hive-shell";
  buildInputs = with pkgs; [ go nixfmt ];
  shellHook = ''
    # We inject <hive> into your path to support local development.
    export NIX_PATH="hive=./nix/hive:$NIX_PATH"
  '';
  # TODO: add chore for:
  #   nixfmt -w 120 *.nix (find example nix -name '*.nix')
}
