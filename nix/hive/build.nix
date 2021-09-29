# <hive/build.nix> builds a named system from the systems section of <deployment> and returns the resulting path.
# This is invoked by Nix-Hive using the paths from <hive/config.nix> like:
#  nix build -I nixpkgs=... -I deployment=... --argstr "name" ... nix/system.nix
{ name }:
let
  deployment = import <deployment>;
  systems = deployment.systems or (throw "systems not specified in deployment");
  system = systems.${name} or (throw "system ${name} could not be found");
  nixos = import <nixpkgs/nixos>;
  result = nixos {
    configuration = if !builtins.hasAttr "configuration" system then
      throw "missing system configuration"
    else if !builtins.isFunction system.configuration then
      throw "system configuration should be a function"
    else
      system.configuration;
    system = system.system or builtins.currentSystem;
  };
in result.system
