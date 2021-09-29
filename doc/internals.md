## How Nix-Hive Evaluates a Deployment

The Nix expression in `<hive/config.nix>` enumerates the deployment configuration but does not build any of
the system configuration paths.  The resulting JSON output contains everything we need to build systems and manage 
instances.  We refer to this output as the system inventory, and it is evaluated each time Nix-Hive is run.

- `paths` -- A list of paths suitable for use with `--include` that will be used when building each system.  This will
  always include `<deployment>`, which will be the path to the deployment.
- `systems.${name}.path` -- A list of paths suitable for use with `--include` when building the system.  These paths 
  will override those in the top level `paths`.
- `instances.${name}.store` -- the instance store that must receive a copy of the system configuration prior to trying
  to transfer it to `${name}`.
- `sshConfig` -- An `ssh_config` (see `man ssh_config`) that contains all of the `instances.${name}.ssh` options.  This
  enables mapping instance names to addresses and specifying jump hosts using `ProxyJump`.  This SSH configuration is
  used with both `nix copy` and running remote commands.
- `systems.${name}.tags` -- A list of tags associated with the system.
- `instances.${instance}.tags` -- A list of tags associated with the instance.

See `go doc . Inventory` for a description of the result's structure structure.

-----------------------
-----------------------
-----------------------

1. `<hive/paths.nix>` is evaluated with `<deployment>` to determine the list of nix include paths.
2. `<hive/systems.nix>` is evaluated with `<deployment>` to determine the systems that should be built.
3. `<hive/system.nix>` is evaluated with each of those systems with the specified nix include paths to produce a
   system configuration path.
4. `<hive/instances.nix>` is evaluated to map the instances to their systems and stores, and build an `ssh_config`.
5. Nix-Hive transfers necessary systems to each store using `nix copy` and the `ssh_config`.
6. Tachimoma transfers a system to each instance using `nix copy`, with substitution from the listed store, if
   specified.
7. Nix-Hive activates each system using the activate script in each system path.

- [ ] `<hive/paths.nix>` and `<hive/systems.nix>` can be done in the same step.

Step 3 must be done one system at a time to support the use of paths, but doing it all together has little real benefit
because Nix does not seem to be significantly faster if you ask it to do it all at once.  Step 4 seems weird, but it
makes the hive.nix's understanding of addresses, names and relationships usable by SSH -- which is important if
you want to use ProxyJump or ProxyCommand to access instances on private networks.

## Building the Inventory from the Deployment Configuration

A Nix-Hive deployment is described by a Nix expression referred to as the "hive.nix" file.  This file contains Nix functions that produce a NixOS system environment when passed to `<nixpkgs/nixos>`.  Nix-Hive uses Nix to build "hive.nix" and convert it into a NixOS configuration using Nix itself.  This lets Nix be at its lazy, memoizing best.  The result is the Nix-Hive inventory, which is a fancy name for a JSON file that identifies each of the instances and their associated configuration.

## Why Does Nix-Hive Use Systems Instead of Per-Instance Configurations?

It takes Nix 5s to determine the top level derivation that describes a trivial system configuration.  If you have 30
instances with 30 identical configurations, it can take 5s to build that system once, or 150s to determine the 
derivation for each individual system.  In short, while Nix is lazy and memoizes the outputs of a derivation in a store,
it does not memoize the derivation itself.

Therefore, we use two tiers -- one that defines the system configurations managed by Nix-Hive, and a second that 
identifies the instances of these systems.

## Alternate Inventory Systems

Nix-Hive can also read the infrastructure to identify instances and their associated systems.  The infrastructure must
provide some form of tags that are used to associate an instance with its system.  This is useful with EC2 deployments
or Terraform.

## Github Deployment Integration

Github has a deployment API which can be used to mark when a deployment begins, finishes, and the outcomes.  This is
useful for providing a central system of record that can be audited.