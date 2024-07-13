{
  description = "web frontend for git";

  inputs.nixpkgs.url = "github:nixos/nixpkgs";

  outputs =
    { self
    , nixpkgs
    ,
    }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in
    {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          legit = self.packages.${system}.legit;
          files = pkgs.lib.fileset.toSource {
            root = ./.;
            fileset = pkgs.lib.fileset.unions [
              ./config.yaml
              ./static
              ./templates
            ];
          };
        in
        {
          legit = pkgs.buildGoModule {
            name = "legit";
            rev = "master";
            src = ./.;

            vendorHash = "sha256-EBVD/RzVpxNcwyVHP1c4aKpgNm4zjCz/99LvfA0Oc/Q=";
          };
          docker = pkgs.dockerTools.buildLayeredImage {
            name = "sini:5000/legit";
            tag = "latest";
            contents = [ files legit pkgs.git ];
            config = {
              Entrypoint = [ "${legit}/bin/legit" ];
              ExposedPorts = { "5555/tcp" = { }; };
            };
          };
        });

      defaultPackage = forAllSystems (system: self.packages.${system}.legit);
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.mkShell {
            nativeBuildInputs = with pkgs; [
              go
            ];
          };
        });
    };
}
