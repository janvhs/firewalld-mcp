// Copyright 2026 SUSE LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/janvhs/firewalld-mcp/internal/firewalld"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ProgramName = "firewalld-mcp"
	MCPName     = "firewalld"
	Version     = "0.1.0"
)

func main() {
	out := flag.CommandLine.Output()

	flag.Parse()

	if err := mainE(); err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
}

func mainE() error {
	ctx := context.Background()

	bus, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("connect system bus: %w", err)
	}
	defer func() { _ = bus.Close() }()

	server := newMCPServer(firewalld.NewDBusClient(bus))

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return err
	}

	return nil
}
