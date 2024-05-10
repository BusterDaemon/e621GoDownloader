{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs }: 
    let
      pkgs = import nixpkgs {
        system = "x86_64-linux";
      };
    in {
      devShells."x86_64-linux".default = pkgs.mkShell {
      name = "E621 & Rule34 downloader";
      hardeningDisable = [ "all" ];
	    buildInputs = with pkgs; [
        go
        gotools
        golangci-lint
        gopls
        go-outline
        gopkgs
        delve
        gcc
	sqlite
	    ];
      shellHook = ''
        go version
      '';
      };
    };
}
