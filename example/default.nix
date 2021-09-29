{
  # Nixpkgs configures the version of NixOS we use for our deployment.  You can override this for each system by 
  # setting the "nixpkgs" attribute for that system.  If this is not specified at either the deployment or system
  # level, <nixpkgs> will be used, which makes the deployment dependent on your environment.
  paths.nixpkgs = builtins.fetchTarball {
    name = "nixpkgs-2021-04-29";
    url = "https://github.com/nixos/nixpkgs/archive/fbadd0efe071f913f9ddd6e0c3b2b1c22d998c14.tar.gz";
    sha256 = "1xjssjlcxxkbyyigy2dvyslk37rp0b5by1q1m1c438g0yypgcpbs";
  };

  # Each system configuration is placed in the system attribute for the deployment.  Unlike other Nix deployment tools,
  # Nix-Hive organizes the system configurations in a distinct section from the instances -- this makes building
  # much faster, since Nix-Hive only needs to build the configuration once per system, and not once per instance.
  systems.www.configuration = { pkgs, ... }: {
    imports = [ ./instance.nix ];
    services.caddy = {
      enable = true;
      config = ''
        example.com {
          file_server /www
        }
      '';
    };
  };

  systems.portico.configuration = import ./instance.nix;

  instances = let
    # internal simplifies describing an internal instance that is not allowed egress or direct contact via ssh.
    internal = system: hostname: {
      inherit system;
      ssh = {
        Hostname = hostname;

        # Our internal hosts can only be accessed by jumping through the "portico" instance.
        ProxyJump = "portico";
      };

      # Since our internal hosts are not allowed egress, their nix.conf uses portico as a substituter.
      # Therefore, we want Nix-Hive to cache any stores there.
      store = "ssh://portico";
    };
  in {
    # portico is an SSH jump box and Nix store used by the www instances.
    portico.system = "portico";

    www-1 = internal "www" "10.22.0.8";
    www-2 = internal "www" "10.22.0.9";
    www-3 = internal "www" "10.22.0.10";
  };
}
