# # See doc/internals.md for an explanation of what this expression does.
let
  inherit (builtins) attrNames concatStringsSep getAttr hasAttr isString mapAttrs;
  deployment = import <deployment>;

  # Explode converts a attrset of name=value pairs into a list of attrsets, each consisting of one pair.
  explode = attrs: map (name: "${name}=${getAttr name attrs}") (attrNames attrs);
  explodePaths = attrs: explode (attrs.paths or { });

  enumerateSystem = name: system: { 
    paths = explodePaths system; 
  };

  # We enumerate all of the systems and their paths.  This serves two functions -- we know which systems need to be
  # built, and we know what paths they have that would override those of the system.
  systems = mapAttrs enumerateSystem (deployment.systems or { });

  # checkSystem checks that a system is a string and identifies a system in systems.
  checkSystem = system:
    if !isString system then
      throw "instance systems must name a system from the systems section"
    else if hasAttr system systems then
      system
    else
      throw "system ${system} not found in the systems section";

  # checkStore checks that a store is a string.
  checkStore = store:
    if !isString store then
      throw ''instance stores must be specified by a URL suitable for use with "nix copy --to"''
    else
      store;

  enumerateInstance = name: instance: {
    tags = instance.tags or [ ];
    store = checkStore (instance.store or "");
    system = checkSystem (instance.system or (throw "instance ${name} does not specify a system."));
  };

  instances = mapAttrs enumerateInstance (deployment.instances or { });

  # We emit the top level paths associated with the deployment, which is provided as the default paths to building
  # the systems.
  paths = (explodePaths deployment);

  sshHostConfig = name:
    let
      config = deployment.instances.${name}.ssh or { };
      options = map (name: "  ${name} ${getAttr name config}") (attrNames config);
    in if options == [ ] then
      ""
    else ''
      Host ${name}
      ${concatStringsSep "\n" options}
    '';

  instanceNames = attrNames instances;
  ssh = concatStringsSep "" (map sshHostConfig instanceNames);
in { inherit instances paths ssh systems; }
