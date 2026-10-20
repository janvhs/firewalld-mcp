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
	"fmt"

	"github.com/janvhs/firewalld-mcp/internal/firewalld"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServer struct {
	client firewalld.FirewalldClient
	server *mcp.Server
}

func newMCPServer(client firewalld.FirewalldClient) *mcpServer {
	s := &mcpServer{
		client: client,
		server: mcp.NewServer(&mcp.Implementation{
			Name:    MCPName,
			Version: Version,
		}, nil),
	}
	s.registerTools()

	return s
}

func (m *mcpServer) Run(ctx context.Context, t mcp.Transport) error {
	return m.server.Run(ctx, t)
}

func (m *mcpServer) Connect(ctx context.Context, t mcp.Transport, opts *mcp.ServerSessionOptions) (*mcp.ServerSession, error) {
	return m.server.Connect(ctx, t, opts)
}

func (m *mcpServer) registerTools() {
	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_default_zone",
		Title:       "Get default zone",
		Description: "Get the name of the default firewall zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *GetDefaultZoneResponse, error) {
		zone, err := m.client.GetDefaultZone(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query default zone: %w", err)
		}
		return nil, &GetDefaultZoneResponse{
			DefaultZone: zone,
		}, nil
	})

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_active_zones",
		Title:       "Get active zones",
		Description: "Get the names of all currently active firewall zones",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ any) (*mcp.CallToolResult, *GetActiveZonesResponse, error) {
		zones, err := m.client.GetActiveZones(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query active zones: %w", err)
		}

		return nil, &GetActiveZonesResponse{
			ActiveZones: zones,
		}, nil
	})

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "get_services_for_zone",
		Title:       "Get services for zone",
		Description: "Get all allowed services of a firewall zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetServicesForZoneArguments) (*mcp.CallToolResult, *GetServicesForZoneResponse, error) {
		services, err := m.client.GetServices(ctx, args.Zone)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query services for zone %q: %w", args.Zone, err)
		}

		return nil, &GetServicesForZoneResponse{
			Services: services,
		}, nil
	})

	mcp.AddTool(m.server, &mcp.Tool{
		Name:        "is_default_zone",
		Title:       "Is default zone",
		Description: "Check whether the specified firewall zone is the default zone",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args IsDefaultZoneArguments) (*mcp.CallToolResult, *IsDefaultZoneResponse, error) {
		defaultZone, err := m.client.GetDefaultZone(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query default zone: %w", err)
		}

		return nil, &IsDefaultZoneResponse{
			IsDefaultZone: args.Zone == defaultZone,
		}, nil
	})

	mcp.AddTool(m.server, &mcp.Tool{
		Name:  "get_service_info",
		Title: "Get service info",
		Description: "Get information on the specified service, including its name, description, ports, " +
			"destination addresses, protocols and source ports. When a service contains items in the includes " +
			"field, you need to request data for each listed include.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GetServiceInfoArguments) (*mcp.CallToolResult, *GetServiceInfoResponse, error) {
		service, err := m.client.GetServiceInfo(ctx, args.Service)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to query info for service %q: %w", args.Service, err)
		}

		return nil, &GetServiceInfoResponse{
			Info: service,
		}, nil
	})
}
