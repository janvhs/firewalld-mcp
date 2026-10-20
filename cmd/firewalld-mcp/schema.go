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

import "github.com/janvhs/firewalld-mcp/internal/firewalld"

type GetDefaultZoneResponse struct {
	DefaultZone string `json:"default_zone"`
}

type GetActiveZonesResponse struct {
	ActiveZones []string `json:"active_zones"`
}

type GetServicesForZoneArguments struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type GetServicesForZoneResponse struct {
	Services []string `json:"services"`
}

type IsDefaultZoneArguments struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type IsDefaultZoneResponse struct {
	IsDefaultZone bool `json:"is_default_zone"`
}

type GetServiceInfoArguments struct {
	Service string `json:"service" jsonschema:"Name of the service"`
}

type GetServiceInfoResponse struct {
	Info Service `json:"info"`
}

type Port struct {
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
}

type Destination struct {
	IPFamily  string `json:"ip_family"`
	IPAddress string `json:"ip_address"`
}

type Service struct {
	Name        string        `json:"name"`
	Short       string        `json:"short"`
	Description string        `json:"description"`
	Ports       []Port        `json:"ports"`
	Destination []Destination `json:"destination"`
	Protocols   []string      `json:"protocols"`
	SourcePorts []Port        `json:"source_ports"`
	Includes    []string      `json:"includes"`
}

func newService(service firewalld.Service) Service {
	return Service{
		Name:        service.Name,
		Short:       service.Short,
		Description: service.Description,
		Ports:       newPorts(service.Ports),
		Destination: newDestinations(service.Destination),
		Protocols:   service.Protocols,
		SourcePorts: newPorts(service.SourcePorts),
		Includes:    service.Includes,
	}
}

func newPorts(ports []firewalld.Port) []Port {
	if ports == nil {
		return nil
	}

	r := make([]Port, 0, len(ports))
	for _, port := range ports {
		r = append(r, Port{
			Port:     port.Port,
			Protocol: port.Protocol,
		})
	}

	return r
}

func newDestinations(destinations []firewalld.Destination) []Destination {
	if destinations == nil {
		return nil
	}

	r := make([]Destination, 0, len(destinations))
	for _, destination := range destinations {
		r = append(r, Destination{
			IPFamily:  destination.IPFamily,
			IPAddress: destination.IPAddress,
		})
	}

	return r
}
