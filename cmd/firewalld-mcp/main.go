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

type GetActiveZonesResp struct {
	ActiveZones []string
}

type GetServicesForZoneArgs struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type GetServicesForZoneResp struct {
	Services []string
}

type IsDefaultZoneArgs struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type IsDefaultZoneResp struct {
	IsDefaultZone bool
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_active_zones",
		Title:       "Get active zones",
		Description: "Get the names of all currently active firewall zones",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *GetActiveZonesResp, error) {
		zones, err := client.GetActiveZones(ctx)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetActiveZonesResp{
			ActiveZones: zones,
		}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_services_for_zone",
		Title:       "Get services for zone",
		Description: "Get all allowed services of a firewall zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetServicesForZoneArgs) (*mcp.CallToolResult, *GetServicesForZoneResp, error) {
		services, err := client.GetServices(ctx, args.Zone)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetServicesForZoneResp{
			Services: services,
		}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "is_default_zone",
		Title:       "Is default zone",
		Description: "Check whether the specified firewall zone is the default zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args IsDefaultZoneArgs) (*mcp.CallToolResult, *IsDefaultZoneResp, error) {
		defaultZone, err := client.GetDefaultZone(ctx)
		if err != nil {
			return nil, nil, err
		}

		return nil, &IsDefaultZoneResp{
			IsDefaultZone: args.Zone == defaultZone,
		}, nil
	})

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return err
	}

	return nil
}
