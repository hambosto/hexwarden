{
  description = "Secure file encryption and decryption CLI tool built with Go";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      ...
    }:
    let
      nixosModule =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.programs.hexwarden;
        in
        {
          options.programs.hexwarden = {
            enable = lib.mkEnableOption "HexWarden Encryption CLI tool";
            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "Package to use for hexwarden";
            };
          };

          config = lib.mkIf cfg.enable {
            environment.systemPackages = [ cfg.package ];
          };
        };

      homeManagerModule =
        {
          config,
          lib,
          pkgs,
          ...
        }:
        let
          cfg = config.programs.hexwarden;
        in
        {
          options.programs.hexwarden = {
            enable = lib.mkEnableOption "HexWarden Encryption CLI tool";
            package = lib.mkOption {
              type = lib.types.package;
              default = self.packages.${pkgs.system}.default;
              description = "Package to use for hexwarden";
            };
          };

          config = lib.mkIf cfg.enable {
            home.packages = [ cfg.package ];
          };
        };

      perSystem =
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          package = pkgs.buildGoModule {
            pname = "hexwarden";
            version = "1.0";
            src = ./.;
            vendorHash = "sha256-KO1Z4XvSvh2B2hgl2DW+WZMXe5gdV7YTYkaATKwtZuM=";
            env.CGO_ENABLED = 0;
            ldflags = [
              "-extldflags '-static'"
              "-s -w"
            ];
          };
        in
        {
          packages.default = package;
        };
    in
    flake-utils.lib.eachDefaultSystem perSystem
    // {
      nixosModules = {
        default = nixosModule;
      };

      homeManagerModules = {
        default = homeManagerModule;
      };
    };
}
