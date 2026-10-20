# Copyright 2026 SUSE LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
{
  description = "Firewalld MCP server";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = inputs: {
    packages =
      builtins.mapAttrs (
        system: pkgs: {
          firewalld-mcp = pkgs.buildGoModule {
            pname = "firewalld-mcp";
            version = "0.1.0";

            src = ./.;

            subPackages = ["cmd/firewalld-mcp"];

            vendorHash = "sha256-Apu/EftiTBBnf/loFGlANWhhs/BLY5oCKBplw1Jhzmg=";

            meta = {
              description = "MCP server to give AI assistants insights into the firewalld configuration";
              homepage = "https://github.com/janvhs/firewalld-mcp";
              license = pkgs.lib.licenses.asl20;
              mainProgram = "firewalld-cmd";
            };
          };

          default = inputs.self.packages.${system}.firewalld-mcp;
        }
      )
      inputs.nixpkgs.legacyPackages;

    devShells =
      builtins.mapAttrs (system: pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            commitlint
            delve
            go
            go-tools
            gopls
            gosec
            gotools
            govulncheck
          ];
        };
      })
      inputs.nixpkgs.legacyPackages;

    formatter =
      builtins.mapAttrs (
        system: pkgs: pkgs.alejandra
      )
      inputs.nixpkgs.legacyPackages;
  };
}
