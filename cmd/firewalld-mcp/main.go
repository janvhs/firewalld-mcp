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
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/janvhs/firewalld-mcp/internal/firewalld"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ProgramName = "firewalld-mcp"
	MCPName     = "firewalld"
	Version     = "0.1.0"
)

type GetDefaultZoneResp struct {
	DefaultZone string
}

func main() {
	flag.Parse()
	if err := mainE(); err != nil {
		log.Fatal(err)
	}
}

func mainE() error {
	ctx := context.Background()

	bus, err := dbus.ConnectSystemBus()
	if err != nil {
		return err
	}
	defer func() { _ = bus.Close() }()

	client := firewalld.NewDBusClient(bus)
	server := mcp.NewServer(&mcp.Implementation{
		Name:    MCPName,
		Version: Version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_default_zone",
		Title:       "Get default zone",
		Description: "Get the name of the default firewall zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *GetDefaultZoneResp, error) {
		zone, err := client.GetDefaultZone(ctx)
		if err != nil {
			return nil, nil, err
		}
		return nil, &GetDefaultZoneResp{
			DefaultZone: string(zone),
		}, nil
	})

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return err
	}

	return nil
}
