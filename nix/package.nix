{ buildGoModule, makeWrapper, stdenv, ... }:
buildGoModule {
  name = "nix-hive";

  src = ./..;
  vendorSha256 = "17sdj4h84q9d22s2j2qx3xrlfpg206x3wlgmq9xfnmk55bzxiidr";
  nativeBuildInputs = [ makeWrapper ];

  postFixup = ''
    mkdir -p $out/nix
    find $out -type f
    cp -r nix/hive/* $out/nix

    # We wrap the Nix-Hive command with a path for <hive> -- the command and its Nix expressions are tightly
    # coupled, so we do not let users override it.
    wrapProgram $out/bin/nix-hive --suffix NIX_PATH : hive=$out/nix
  '';

  #TODO: Add meta, version.
}
