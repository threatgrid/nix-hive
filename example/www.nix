{ pkgs, ... }: {
  imports = [ ./instance.nix ];
  services.caddy = {
    enable = true;
    config = ''
      example.com {
        file_server /www
      }
    '';
  };
}
