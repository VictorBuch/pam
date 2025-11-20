args@{
  config,
  pkgs,
  lib,
  inputs ? null,
  isLinux,
  mkApp,
  ...
}:

mkApp {
  _file = toString ./.;
  name = "PackageName";
  description = "PackageDescription";
  linuxPackages = pkgs: [ LinuxPackage ];
  darwinPackages = pkgs: [ DarwinPackage ];
  darwinExtraConfig = { homebrew.casks = [ "HomebrewPackage" ]; };
} args
