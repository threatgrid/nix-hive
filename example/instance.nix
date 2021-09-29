{ pkgs, ... }: {
  ec2.hvm = true;
  imports = [ <nixpkgs/nixos/modules/virtualisation/amazon-image.nix> ];
}
