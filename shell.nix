{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  packages = with pkgs; [
    go
    gopls
    gotests
    go-tools
    protoc-gen-go
    protoc-gen-go-grpc
    protobuf
    openssl
  ];
}