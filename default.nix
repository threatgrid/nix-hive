# This Nix expression enables installation from a URL or local checkout of Hive:
#
# From URL: nix-env -if https://github.com/threatgrid/nix-hive/archive/main.tar.gz
# From the current directory: nix-env -if .
# 
# There is also an overlay in nix/overlay.nix, a simple package in nix/package.nix and a nix shell in shell.nix.
{ nixpkgs ? <nixpkgs> }:
let pkgs = import nixpkgs { overlays = [ (import nix/overlay.nix) ]; };
in pkgs.hive
