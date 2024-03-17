{
  pkgs,
}:
 pkgs.buildGoModule {
  pname = "givc-agent";
  version = "0.0.1";
  src = ../../.;
  vendorHash = "sha256-s1UOASsmXUaZmCZMos9KvXMu8f7iyUxQHlRe5Tc4T3g=";
  subPackages = [
    "api/admin"
    "api/systemd"
    "internal/pkgs/grpc"
    "internal/pkgs/servicemanager"
    "internal/pkgs/serviceclient"
    "internal/pkgs/utility"
    "internal/cmd/givc-agent"
  ];
}