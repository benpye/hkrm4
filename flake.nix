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
          version = "2022-01-27";
          vendorSha256 = "sha256-8mGrTvrJS16W8hvgpdewFNvTt+p4Tr0PgHZPKksmfZc=";
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
