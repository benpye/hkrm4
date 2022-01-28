{
  description = "A flake for building hkrm4.";

  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    with import nixpkgs;
    let
      name = "hkrm4";
      src = self;
    in
    {
      overlay = self: super: {
        ${name} = super.buildGoModule {
          inherit name src;
          version = "2022-01-28";
          vendorSha256 = "sha256-w5VEKWdY+kQyWEvKRZ6FovfzpwRW+zt2lFKylndvhU4=";
          subPackages = [ "cmd/hkrm4" ];
        };
      };
    } // (
      flake-utils.lib.eachDefaultSystem (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ self.overlay ];
          };

          package = pkgs.${name};
        in {
          packages.${name} = package;
          defaultPackage = package;
        }
      )
    );
}
