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

// Package firewalld is a client for the firewalld D-Bus API.
package firewalld

import (
	"context"
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

// ErrIPC is the error returned when the communication to D-Bus failed.
var ErrIPC = errors.New("firewalld: failed D-Bus communication")

const (
	dbusDest              = "org.fedoraproject.FirewallD1"
	dbusPath              = "/org/fedoraproject/FirewallD1"
	dbusFirewallInterface = "org.fedoraproject.FirewallD1"
	dbusZoneInterface     = "org.fedoraproject.FirewallD1.zone"
)

type Port struct {
	Port     string
	Protocol string
}

type Destination struct {
	IPFamily  string
	IPAddress string
}

type Service struct {
	Name        string
	Short       string
	Description string
	Ports       []Port
	Destination []Destination
	Protocols   []string
	SourcePorts []Port
	Includes    []string
}

type FirewalldClient interface {
	GetDefaultZone(context.Context) (string, error)
	GetActiveZones(context.Context) ([]string, error)
	GetServices(context.Context, string) ([]string, error)
	GetServiceInfo(context.Context, string) (Service, error)
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

	// getDefaultZone returns s.
	var zone string
	if err := c.object.CallWithContext(ctx, method, 0).Store(&zone); err != nil {
		return "", fmt.Errorf("%w: %w", ErrIPC, err)
	}

	return zone, nil
}

func (c *dbusClient) GetActiveZones(ctx context.Context) ([]string, error) {
	method := dbusZoneInterface + ".getActiveZones"

	// getActiveZones returns a{sa{sas}}, mapping each active zone name to
	// its assigned interfaces and sources. Only the zone names are needed.
	var activeZones map[string]any
	if err := c.object.CallWithContext(ctx, method, 0).Store(&activeZones); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIPC, err)
	}

	zones := make([]string, 0, len(activeZones))
	for zone := range activeZones {
		zones = append(zones, zone)
	}

	return zones, nil
}

func (c *dbusClient) GetServices(ctx context.Context, zone string) ([]string, error) {
	method := dbusZoneInterface + ".getServices"

	// getServices returns as.
	var services []string
	if err := c.object.CallWithContext(ctx, method, 0, zone).Store(&services); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIPC, err)
	}

	return services, nil
}

func (c *dbusClient) GetServiceInfo(ctx context.Context, service string) (Service, error) {
	method := dbusFirewallInterface + ".getServiceSettings2"

	// getServiceSettings2 returns a{sv}.
	var settings map[string]dbus.Variant
	if err := c.object.CallWithContext(ctx, method, 0, service).Store(&settings); err != nil {
		return Service{}, fmt.Errorf("%w: %w", ErrIPC, err)
	}

	return serviceFromSettings(service, settings)
}

func serviceFromSettings(name string, settings map[string]dbus.Variant) (Service, error) {
	info := Service{Name: name}

	var err error
	if info.Short, err = storeVariantMap(settings, "short", ""); err != nil {
		return Service{}, err
	}
	if info.Description, err = storeVariantMap(settings, "description", ""); err != nil {
		return Service{}, err
	}
	if info.Ports, err = storeVariantMap(settings, "ports", []Port{}); err != nil {
		return Service{}, err
	}
	if info.Protocols, err = storeVariantMap(settings, "protocols", []string{}); err != nil {
		return Service{}, err
	}
	if info.SourcePorts, err = storeVariantMap(settings, "source_ports", []Port{}); err != nil {
		return Service{}, err
	}
	if info.Includes, err = storeVariantMap(settings, "includes", []string{}); err != nil {
		return Service{}, err
	}

	rawDestination, err := storeVariantMap(settings, "destination", map[string]string{})
	if err != nil {
		return Service{}, err
	}
	destination := make([]Destination, 0)
	for family, address := range rawDestination {
		destination = append(destination, Destination{
			IPFamily:  family,
			IPAddress: address,
		})
	}
	info.Destination = destination

	return info, nil
}

func storeVariantMap[T any](m map[string]dbus.Variant, key string, fallback T) (T, error) {
	variant, ok := m[key]
	if !ok {
		return fallback, nil
	}

	var v T
	if err := dbus.Store([]any{variant.Value()}, &v); err != nil {
		return fallback, fmt.Errorf("%w: firewalld: unexpected D-Bus signature %q for field %q: %w", ErrIPC, variant.Signature(), key, err)
	}

	return v, nil
}
