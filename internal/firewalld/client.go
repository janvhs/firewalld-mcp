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
package firewalld

import (
	"context"

	"github.com/godbus/dbus/v5"
)

const (
	dbusDest              = "org.fedoraproject.FirewallD1"
	dbusPath              = "/org/fedoraproject/FirewallD1"
	dbusFirewallInterface = "org.fedoraproject.FirewallD1"
	dbusZoneInterface     = "org.fedoraproject.FirewallD1.zone"
)

type Port struct {
}

type Destination struct {
}

type Service struct {
}

type FirewalldClient interface {
	GetDefaultZone(context.Context) (string, error)
	GetActiveZones(context.Context) ([]string, error)
	GetServices(context.Context, string) ([]string, error)
}

var _ FirewalldClient = (*dbusClient)(nil)

type dbusClient struct {
	bus    *dbus.Conn
	object dbus.BusObject
}

func NewDBusClient(bus *dbus.Conn) FirewalldClient {
	return &dbusClient{
		bus:    bus,
		object: bus.Object(dbusDest, dbusPath),
	}
}

func (c *dbusClient) GetDefaultZone(ctx context.Context) (string, error) {
	method := dbusFirewallInterface + ".getDefaultZone"

	var zone string
	if err := c.object.CallWithContext(ctx, method, 0).Store(&zone); err != nil {
		return "", err
	}

	return zone, nil
}

func (c *dbusClient) GetActiveZones(ctx context.Context) ([]string, error) {
	method := dbusZoneInterface + ".getActiveZones"

	// getActiveZones returns a{sa{sas}}, mapping each active zone name to
	// its assigned interfaces and sources. Only the zone names are needed.
	var activeZones map[string]any
	if err := c.object.CallWithContext(ctx, method, 0).Store(&activeZones); err != nil {
		return nil, err
	}

	zones := make([]string, 0, len(activeZones))
	for zone := range activeZones {
		zones = append(zones, zone)
	}

	return zones, nil
}

func (c *dbusClient) GetServices(ctx context.Context, zone string) ([]string, error) {
	method := dbusZoneInterface + ".getServices"

	var services []string
	if err := c.object.CallWithContext(ctx, method, 0, zone).Store(&services); err != nil {
		return nil, err
	}

	return services, nil
}
