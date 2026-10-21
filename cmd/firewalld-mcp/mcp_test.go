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
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/janvhs/firewalld-mcp/internal/firewalld"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mockClient struct {
	activeZones []string
	defaultZone string
	serviceInfo map[string]firewalld.Service
	services    map[string][]string
	err         error
}

var _ firewalld.FirewalldClient = (*mockClient)(nil)

func (c *mockClient) GetDefaultZone(context.Context) (string, error) {
	if c.err != nil {
		return "", c.err
	}

	return c.defaultZone, nil
}

func (c *mockClient) GetActiveZones(context.Context) ([]string, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.activeZones, nil
}

func (c *mockClient) GetServices(_ context.Context, zone string) ([]string, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.services[zone], nil
}

func (c *mockClient) GetServiceInfo(_ context.Context, service string) (firewalld.Service, error) {
	if c.err != nil {
		return firewalld.Service{}, c.err
	}

	return c.serviceInfo[service], nil
}

func newMockClient() *mockClient {
	return &mockClient{
		defaultZone: "public",
		activeZones: []string{"empty", "public"},
		services: map[string][]string{
			"public": {"dhcpv6-client", "ssh"},
			"empty":  {},
		},
		serviceInfo: map[string]firewalld.Service{
			"ssh": {
				Name:        "ssh",
				Short:       "SSH",
				Description: "Secure Shell (SSH) is a protocol for logging into and executing commands on remote machines. It provides secure encrypted communications. If you plan on accessing your machine remotely via SSH over a firewalled interface, enable this option. You need the openssh-server package installed for this option to be useful.",
				Ports:       []firewalld.Port{{Port: "22", Protocol: "tcp"}},
				Destination: []firewalld.Destination{},
				Protocols:   []string{},
				SourcePorts: []firewalld.Port{},
				Includes:    []string{},
			},
		},
	}
}

func connectMCP(t *testing.T, client firewalld.FirewalldClient) (context.Context, *mcp.ClientSession) {
	t.Helper()

	ctx := t.Context()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	if _, err := newMCPServer(client).Connect(ctx, serverTransport, nil); err != nil {
		t.Fatalf("mcpServer.Connect() returned an error: %v", err)
	}

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil).
		Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("mcp.Client.Connect() returned an error: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	return ctx, session
}

func callTool[T any](t *testing.T, ctx context.Context, session *mcp.ClientSession, name string, args map[string]any) T {
	t.Helper()

	resp, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("session.CallTool(%q) returned an error: %v", name, err)
	}
	if resp.IsError {
		t.Fatalf("session.CallTool(%q) is an error: %+v", name, resp.Content)
	}

	raw, err := json.Marshal(resp.StructuredContent)
	if err != nil {
		t.Fatalf("json.Marshal(%q -> resp) returned an error: %v", name, err)
	}

	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("json.Unmarshal(%q -> raw) returned an error: %v", name, err)
	}

	return v
}

func TestToolsAreRegistered(t *testing.T) {
	want := []string{
		"get_active_zones",
		"get_default_zone",
		"get_service_info",
		"get_services_for_zone",
		"is_default_zone",
	}

	ctx, session := connectMCP(t, newMockClient())

	var got []string
	for tool, err := range session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("session.Tools() returned an error: %v", err)
		}
		got = append(got, tool.Name)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("values don't match (-want +got):\n%s", diff)
	}
}

func TestGetActiveZones(t *testing.T) {
	client := newMockClient()
	want := client.activeZones

	ctx, session := connectMCP(t, client)
	got := callTool[GetActiveZonesResponse](t, ctx, session, "get_active_zones", nil)
	if diff := cmp.Diff(want, got.ActiveZones); diff != "" {
		t.Errorf("values don't match (-want +got):\n%s", diff)
	}
}

func TestGetDefaultZone(t *testing.T) {
	client := newMockClient()
	want := client.defaultZone

	ctx, session := connectMCP(t, client)
	got := callTool[GetDefaultZoneResponse](t, ctx, session, "get_default_zone", nil)
	if got.DefaultZone != want {
		t.Errorf("got %q, want %q", got.DefaultZone, want)
	}
}

func TestGetServicesForZone(t *testing.T) {
	tests := map[string]struct {
		name string
		zone string
		want []string
	}{
		"zone with services":    {zone: "public", want: []string{"dhcpv6-client", "ssh"}},
		"zone without services": {zone: "empty", want: []string{}},
		"nonexistent zone":      {zone: "invalid", want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, session := connectMCP(t, newMockClient())

			got := callTool[GetServicesForZoneResponse](t, ctx, session, "get_services_for_zone",
				map[string]any{"zone": test.zone})
			if diff := cmp.Diff(test.want, got.Services); diff != "" {
				t.Errorf("values don't match (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsDefaultZone(t *testing.T) {
	tests := []struct {
		zone string
		want bool
	}{
		{zone: "public", want: true},
		{zone: "empty", want: false},
		{zone: "", want: false},
	}

	for _, test := range tests {
		t.Run(test.zone, func(t *testing.T) {
			ctx, session := connectMCP(t, newMockClient())

			got := callTool[IsDefaultZoneResponse](t, ctx, session, "is_default_zone",
				map[string]any{"zone": test.zone})
			if got.IsDefaultZone != test.want {
				t.Errorf("IsDefaultZone(%q) = %v, want %v", test.zone, got.IsDefaultZone, test.want)
			}
		})
	}
}

func TestGetServiceInfo(t *testing.T) {
	want := Service{
		Name:        "ssh",
		Short:       "SSH",
		Description: "Secure Shell (SSH) is a protocol for logging into and executing commands on remote machines. It provides secure encrypted communications. If you plan on accessing your machine remotely via SSH over a firewalled interface, enable this option. You need the openssh-server package installed for this option to be useful.",
		Ports:       []Port{{Port: "22", Protocol: "tcp"}},
		Destination: []Destination{},
		Protocols:   []string{},
		SourcePorts: []Port{},
		Includes:    []string{},
	}

	ctx, session := connectMCP(t, newMockClient())
	got := callTool[GetServiceInfoResponse](t, ctx, session, "get_service_info",
		map[string]any{"service": "ssh"})

	if diff := cmp.Diff(want, got.Info); diff != "" {
		t.Errorf("values don't match (-want +got):\n%s", diff)
	}
}

func TestToolErrorsAreReported(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
	}{
		{name: "get_default_zone"},
		{name: "get_active_zones"},
		{name: "get_services_for_zone", args: map[string]any{"zone": "public"}},
		{name: "is_default_zone", args: map[string]any{"zone": "public"}},
		{name: "get_service_info", args: map[string]any{"service": "ssh"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, session := connectMCP(t, &mockClient{err: errors.New("test error")})

			res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: test.name, Arguments: test.args})
			if err != nil {
				return
			}
			if !res.IsError {
				t.Errorf("call %s succeeded, want an error result", test.name)
			}
		})
	}
}
