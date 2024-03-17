{
  pkgs,
}:
 pkgs.buildGoModule {
  pname = "givc-admin";
  version = "0.0.1";
  src = ../../.;
  vendorHash = "sha256-s1UOASsmXUaZmCZMos9KvXMu8f7iyUxQHlRe5Tc4T3g=";
  subPackages = [
    "api/admin"
    "api/systemd"
    "internal/pkgs/grpc"
    "internal/pkgs/registry"
    "internal/pkgs/systemmanager"
    "internal/pkgs/serviceclient"
    "internal/pkgs/types"
    "internal/pkgs/utility"
    "internal/cmd/givc-admin"
  ];

}
