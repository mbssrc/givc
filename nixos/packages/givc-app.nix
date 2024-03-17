{
  pkgs,
}:
 pkgs.buildGoModule {
  pname = "givc-app";
  version = "0.0.1";
  src = ../../.;
  vendorHash = "sha256-s1UOASsmXUaZmCZMos9KvXMu8f7iyUxQHlRe5Tc4T3g=";
  subPackages = [
    "api/admin"
    "internal/pkgs/grpc"
    "internal/pkgs/types"
    "internal/pkgs/utility"
    "internal/cmd/givc-app"
  ];
}
