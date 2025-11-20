{ config, pkgs, ... }:

{
  imports = [ ./hardware-configuration.nix ];

  apps = {
    browsers = {
      firefox.enable = true;
      chrome.enable = false;
    };

    editors = {
      neovim.enable = true;
    };
  };

  system.stateVersion = "23.11";
}
