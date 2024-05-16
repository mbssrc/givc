{
  description = "Go modules for inter-vm communication with gRPC.";

  # Inputs
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    devshell.url = "github:numtide/devshell";
  };

  outputs = inputs@{ self, flake-utils, devshell, nixpkgs }:
    let
      # Generate a user-friendly version number
      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";
      version = builtins.substring 0 8 lastModifiedDate;

      systems = with flake-utils.lib.system; [
        x86_64-linux
        aarch64-linux
      ];
    in
      flake-utils.lib.eachSystem systems (system: {

        # Packages
        packages =
          let
            pkgs = nixpkgs.legacyPackages.${system};
          in {
            default = pkgs.callPackage ./nixos/packages/default.nix {};
            givc-app = pkgs.callPackage ./nixos/packages/givc-app.nix {};
          };

        # DevShells
        devShells =
          let
            pkgs = nixpkgs.legacyPackages.${system}.extend( devshell.overlays.default );
          in {
            default = pkgs.devshell.mkShell {
              imports = [ (pkgs.devshell.importTOML ./devshell.toml) ];
            };
          };

      }) // {

        # NixOS Modules
        nixosModules =
          {
            admin = import ./nixos/modules/admin.nix;
            host = import ./nixos/modules/host.nix;
            sysvm = import ./nixos/modules/sysvm.nix;
            appvm = import ./nixos/modules/appvm.nix;
          };

        # Overlays
        overlays.default = (final: prev:{
          givc-app = prev.callPackage ./nixos/packages/givc-app.nix { pkgs=prev; };
        });
      };
}
