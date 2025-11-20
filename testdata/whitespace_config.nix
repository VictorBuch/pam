{ config, pkgs, ... }:

{
  imports = [ ./hardware-configuration.nix ];

  apps={
    browsers  =  {
      firefox.enable = true;
    };
  };

  system.stateVersion = "23.11";
}
