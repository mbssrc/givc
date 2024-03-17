{
  pkgs,
}:
let
  version = "0.0.1";
  rev = "356a1ecc3efdcfcda684401eb8788c40c0e90c3d";
in
 pkgs.buildGoModule {
  pname = "givc";
  inherit version;

  src = pkgs.fetchFromGitHub {
    owner = "mbssrc";
    repo = "givc";
    inherit rev;
    hash = "sha256-mlAx1z8zio3qCvCmiKTVIAJQBzuAjh6T/Y4P1cVQvuU=";
  };
  vendorHash = "sha256-s1UOASsmXUaZmCZMos9KvXMu8f7iyUxQHlRe5Tc4T3g=";
}
