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
	DefaultZone string `json:"default_zone" jsonschema:"Name of the default firewall zone"`
}

type GetActiveZonesResponse struct {
	ActiveZones []string `json:"active_zones" jsonschema:"Names of all currently active firewall zones"`
}

type GetServicesForZoneArguments struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type GetServicesForZoneResponse struct {
	Services []string `json:"services" jsonschema:"Names of the services allowed in the zone"`
}

type IsDefaultZoneArguments struct {
	Zone string `json:"zone" jsonschema:"Name of the firewall zone"`
}

type IsDefaultZoneResponse struct {
	IsDefaultZone bool `json:"is_default_zone" jsonschema:"Whether the requested zone is the default zone"`
}

type GetServiceInfoArguments struct {
	Service string `json:"service" jsonschema:"Name of the service"`
}

type GetServiceInfoResponse struct {
	Info Service `json:"info" jsonschema:"Definition of the requested service"`
}

type Port struct {
	Port     string `json:"port" jsonschema:"Port number or port range such as 22 or 6800-7568"`
	Protocol string `json:"protocol" jsonschema:"Transport protocol of the port. The value can either be tcp, udp, sctp or dccp"`
}

type Destination struct {
	IPFamily  string `json:"ip_family" jsonschema:"IP family of the destination network. The value can either be ipv4 or ipv6"`
	IPAddress string `json:"ip_address" jsonschema:"Address of the destination network. Can be a plain IP address or network IP address (optional with /mask). Hostnames are allowed but discouraged, since they are resolved only when the service is activated"`
}

type Service struct {
	Name        string        `json:"name" jsonschema:"Name the service is known by to firewalld"`
	Short       string        `json:"short" jsonschema:"Human readable name of the service. Show this when talking about a service"`
	Description string        `json:"description" jsonschema:"Description of the service"`
	Ports       []Port        `json:"ports" jsonschema:"Ports the service matches. An entry with an empty port only matches the protocol"`
	Destination []Destination `json:"destination" jsonschema:"Destination network the service matches. Analogous to --destination in iptables and ip6tables. At most one entry per IP family"`
	Protocols   []string      `json:"protocols" jsonschema:"Protocols the service matches. Values come from /etc/protocols and can be specified by name or number"`
	SourcePorts []Port        `json:"source_ports" jsonschema:"Ports the service matches as the source port"`
	Includes    []string      `json:"includes" jsonschema:"Further services the service includes. Request each of them separately"`
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
