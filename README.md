# Nix-Hive, Building Identical Systems From a Single Source With NixOS

Nix-Hive is a utility for managing a large number of NixOS hosts with identical configurations, leveraging Nix stores,
substitutions and path signing to simplify maintenance.  This tool evolved from tools used to manage internal ThreatGRID 
projects and is provided in the hopes that it will be useful to others, without warranty or promise of support.  There 
are a number of similar tools out there, but they all lacked one or more of the following:

- No need for a shared state database.  State is determined from the deployment configuration and the hosts.
- Strong support for the use of Nix stores for caching and transmission.  Deployments may involve thousands of
  instances, pushing from the build instance is not appropriate.
- No need to build configurations on the deployment instance.  Instances can be very small and do not need access to
  internet sources.
- SSH configuration derived from deployment configuration to simplify proxy jumping, key fingerprints and host name
  resolution.
- Focus on building a small set of systems that are deployed to a large number of instances.  This prevents repeated
  builds of identical configurations that consumes considerable time on the build host.

## Installation and Requirements

Nix-Hive depends on Nix to build systems and transfer them, and therefore can be installed using Nix:

```sh
nix-env -if https://github.com/threatgrid/nix-hive/archive/main.tar.gz
```

The build host should be the same system type as the managed instances.  (Meaning, you cannot currently use Nix-Hive
on a Mac to manage Linux instances, or an ARM host to manage amd64 instances.)

## A Simple Example

The [example](./example) subdirectory contains an example of deploying a jumpbox, web server and internal storage
server.  The parts specific to Nix-Hive are found in [default.nix](./example/default.nix), while the other files are
NixOS configuration modules, identical to what you would expect from /etc/nixos/configuration.nix.

A Nix-Hive "hive.nix" contains a Nix expression that maps these configuration modules into named "systems" that
Nix-Hive will build and activate on "instances."  It may also include information about Nix include paths used to
build the systems, which may be overridden on a per-system basis.

Deployment of the example should be as simple as updating the addresses and running:

```
nix-hive deploy -c example portico '*'
```

You should call out "portico" first, before adding a wildcard to match all of the instances to ensure that portico is 
updated first.  In this deployment, portico is the jump box, and the web instances are not directly accessible over
the internet.  Nix-Hive is smart enough to identify that Portico is targeted twice, once by name, and again by the
wildcard, and will only deploy to it once.  It will also push the web system to Portico first, so it is available to the
web instances, which use Portico as a store.

In response, Nix-Hive will:

- Build each of the systems required by the targeted instances.  In our example, that is "portico" and "www"
- Transfer each of the systems to "portico."  The "www" instances specify that their store is "portico", so that
  instance will receive both its own system and the ones needed by www.
- Transfer the systems to the three "www" servers.  These servers are configured to use portico as a substituter, so
  only the path is transferred, not the contents of the path or its dependencies.
- Activate the systems on each host, first on Portico, then on WWW.

Running this command again will cause Nix-Hive to rebuild the systems, because it does not know if there was a change
it could not detect, then transfer and activate them.  Due to how NixOS systems work, little to nothing will change on
each of the instances -- activating the same system twice does nothing.

## Caching State

While Nix-Hive deliberately does not depend on a centralized store for state, it is useful to skip steps when an issue 
occurs, such as a host being temporarily unavailable and needing redeployment.  Nix-Hive caches state in a file, 
`hive.state`, that can be used to skip steps, using the `--no` flag to specify which steps it can skip.  Without
using `--no`, Nix-Hive will follow all the steps, every time, even if it has a `hive.state` file.

Steps Nix-Hive can skip:

- `build` -- Do not build a system if its previous build was saved in local state and is in the Nix store.

## Secret Management

Nix-Hive currently has no facilities for managing secrets.  You should never store a secret in the Nix store, since
the store must be readable by all processes on the host.  Other tools like Nix-Hive do clever things using tmpfs and
scp.  (We use a dedicated secrets manager as a part of our infrastructure.)

## Instance Discovery

Nix-Hive does not discover instances itself.  Instead, you should use a tool like `pulumi` or `terraform` that 
generates a file identifying the instances and their associated systems, and import it into your `hive.nix`.

## Remote Build

Nix-Hive currently does not support using a build host to build its systems.  A pull request adding this would be very
welcome, but it is not something we needed for our own use.

## Running Tasks

Nix-Hive can construct an SSH configuration from information in the `hive.nix` and exposes this configuration
with the `nix-hive ssh` and `nix-hive scp` subcommands.  This configuration will extend your own `.ssh/config` and
`.ssh/known_hosts`.

You can also use `nix-hive run` to run a command on multiple instances.

## Nix Channels

NixOS provides distinct Nixpkg "channels" for managing the update frequency and stability of Nix configurations.  While
Nix-Hive does not support this directly, you can wire the URLs for the channels into `paths.nixpkgs` at either the 
deployment or system level.

## Anticipated Questions

### How can I list my systems or instances?

The `nix-hive build` command outputs a JSON object which describes the inventory as Nix Hive understands it.  If you add the `--no build` flag, `nix-hive build` will skip actually building your systems, making this faster.

```
nix-hive build --no build | jq .
```

## Version History

### 0.1 -- First Public Version of Nix Hive

Nix Hive is a tool for building and deploying Nix systems to multiple hosts in an efficient way.  It is derived from
our existing tools and made available with no warranty and plenty of a good will.