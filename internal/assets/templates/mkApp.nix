{ lib }:

# Universal app module helper with cross-platform and stable/unstable support

{
  name,
  # Auto-derive optionPath from file location (or specify manually)
  _file ? null,
  optionPath ? null,
  # Cross-platform (use same packages for both)
  packages ? null,
  # Platform-specific packages
  linuxPackages ? packages,
  darwinPackages ? packages,
  # Description
  description ? "Enables ${name}",
  # Extra config (applies to both platforms)
  extraConfig ? { },
  # Platform-specific extra config
  linuxExtraConfig ? { },
  darwinExtraConfig ? { },
}:

{
  config,
  pkgs,
  lib,
  inputs ? null,
  isLinux ? pkgs.stdenv.isLinux,
  ...
}:

let
  # Auto-derive optionPath from file location if not explicitly provided
  derivedOptionPath =
    if optionPath != null then
      optionPath
    else if _file != null then
      let
        # Convert file path to string and extract the part after "modules/apps/"
        filePath = toString _file;
        # Split by "modules/apps/" and take the part after it
        parts = lib.splitString "modules/apps/" filePath;
        # Get the relative path (e.g., "browsers" or "gaming/utils")
        relativePath = if (builtins.length parts) > 1 then builtins.elemAt parts 1 else "";
        # Convert path separators to dots (e.g., "gaming/utils" -> "gaming.utils")
        categoryPath = builtins.replaceStrings [ "/" ] [ "." ] relativePath;
        # Build the full option path: "apps.category.name"
        fullPath = "apps.${categoryPath}.${name}";
      in
      fullPath
    else
      throw "mkApp: Either '_file' or 'optionPath' must be provided for ${name}";

  # Import stable nixpkgs if inputs is available
  stable-pkgs =
    if inputs != null then
      import inputs.stable-nixpkgs {
        inherit (pkgs) system;
        config.allowUnfree = true;
      }
    else
      pkgs;

  # Select platform-specific packages
  platformPackages = if isLinux then linuxPackages else darwinPackages;

  # Resolve packages (support functions or direct lists)
  resolvePackages =
    pkgList:
    if pkgList == null then
      [ ]
    else if lib.isFunction pkgList then
      let
        funcArgs = lib.functionArgs pkgList;
      in
      if funcArgs != { } then
        # Named arguments: { pkgs, stable-pkgs }
        pkgList { inherit pkgs stable-pkgs; }
      else
        # Single argument: pkgs
        pkgList pkgs
    else
      # Direct list
      pkgList;

  resolvedPackages = resolvePackages platformPackages;

  # Detect if this is platform-specific
  isLinuxOnly = linuxPackages != null && darwinPackages == null;
  isDarwinOnly = darwinPackages != null && linuxPackages == null;

  # Merge platform-specific and common extra config
  platformExtraConfig = if isLinux then linuxExtraConfig else darwinExtraConfig;
  finalExtraConfig = lib.recursiveUpdate extraConfig platformExtraConfig;

  # Build the option path dynamically from derived or provided optionPath
  optionParts = lib.splitString "." derivedOptionPath;
in

{
  options = lib.setAttrByPath (optionParts ++ [ "enable" ]) (lib.mkEnableOption description);

  config =
    let
      # Use attrByPath with default to avoid errors when option doesn't exist yet
      optionEnabled = lib.attrByPath optionParts { enable = false; } config;
    in
    lib.mkMerge [
      # Apply configuration when enabled AND platform is compatible
      # Silently skip if platform is incompatible (no error, just don't enable)
      (lib.mkIf
        (
          optionEnabled.enable
          &&
            # Skip if Linux-only and we're on Darwin
            !(isLinuxOnly && !isLinux)
          &&
            # Skip if Darwin-only and we're on Linux
            !(isDarwinOnly && isLinux)
        )
        (
          lib.recursiveUpdate {
            environment.systemPackages = resolvedPackages;
          } finalExtraConfig
        )
      )
    ];
}
