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
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func portVariant(pairs ...[2]string) dbus.Variant {
	ports := make([][]any, 0, len(pairs))
	for _, p := range pairs {
		ports = append(ports, []any{p[0], p[1]})
	}

	return dbus.MakeVariant(ports)
}

func TestServiceFromSettings(t *testing.T) {
	tests := map[string]struct {
		service  string
		settings map[string]dbus.Variant
		want     Service
	}{
		"empty settings fall back to empty values": {
			service:  "ssh",
			settings: map[string]dbus.Variant{},
			want: Service{
				Name:        "ssh",
				Short:       "",
				Description: "",
				Ports:       []Port{},
				Destination: []Destination{},
				Protocols:   []string{},
				SourcePorts: []Port{},
				Includes:    []string{},
			},
		},
		"most common shape is accepted": {
			service: "ssh",
			settings: map[string]dbus.Variant{
				"short":       dbus.MakeVariant("SSH"),
				"description": dbus.MakeVariant("Secure Shell (SSH) is a protocol for logging into and executing commands on remote machines. It provides secure encrypted communications. If you plan on accessing your machine remotely via SSH over a firewalled interface, enable this option. You need the openssh-server package installed for this option to be useful."),
				"ports":       portVariant([2]string{"22", "tcp"}),
			},
			want: Service{
				Name:        "ssh",
				Short:       "SSH",
				Description: "Secure Shell (SSH) is a protocol for logging into and executing commands on remote machines. It provides secure encrypted communications. If you plan on accessing your machine remotely via SSH over a firewalled interface, enable this option. You need the openssh-server package installed for this option to be useful.",
				Ports:       []Port{{Port: "22", Protocol: "tcp"}},
				Destination: []Destination{},
				Protocols:   []string{},
				SourcePorts: []Port{},
				Includes:    []string{},
			},
		},
		"empty port is accepted": {
			service: "ipsec",
			settings: map[string]dbus.Variant{
				"short":       dbus.MakeVariant("IPsec"),
				"description": dbus.MakeVariant("Internet Protocol Security (IPsec) is the standardized IETF VPN architecture defined in RFC 4301. IPsec is negotiated using the IKEv1 (RFC 2409) or IKEv2 (RFC 7296) protocol, which in itself uses encryption and authentication. IPsec provides Internet Protocol (IP) packet encryption and authentication. Both IKE and IPsec can be encapsulated in UDP (RFC 3948) or TCP (RFC 8229 to make it easier to traverse NAT. Enabling this service will enable IKE, IPsec and their encapsulation protocols and ports. Note that IKE and IPsec can also be configured to use non-default ports, but this is not common practice."),
				"ports": portVariant(
					[2]string{"", "ah"},
					[2]string{"", "esp"},
					[2]string{"500", "udp"},
					[2]string{"4500", "udp"},
					[2]string{"4500", "tcp"},
				),
			},
			want: Service{
				Name:        "ipsec",
				Short:       "IPsec",
				Description: "Internet Protocol Security (IPsec) is the standardized IETF VPN architecture defined in RFC 4301. IPsec is negotiated using the IKEv1 (RFC 2409) or IKEv2 (RFC 7296) protocol, which in itself uses encryption and authentication. IPsec provides Internet Protocol (IP) packet encryption and authentication. Both IKE and IPsec can be encapsulated in UDP (RFC 3948) or TCP (RFC 8229 to make it easier to traverse NAT. Enabling this service will enable IKE, IPsec and their encapsulation protocols and ports. Note that IKE and IPsec can also be configured to use non-default ports, but this is not common practice.",
				Ports: []Port{
					{Port: "", Protocol: "ah"},
					{Port: "", Protocol: "esp"},
					{Port: "500", Protocol: "udp"},
					{Port: "4500", Protocol: "udp"},
					{Port: "4500", Protocol: "tcp"},
				},
				Destination: []Destination{},
				Protocols:   []string{},
				SourcePorts: []Port{},
				Includes:    []string{},
			},
		},
		"destination and protocols are accepted": {
			service: "vrrp",
			settings: map[string]dbus.Variant{
				"short":       dbus.MakeVariant("Virtual Router Redundancy Protocol RFC3768 and RFC5798."),
				"description": dbus.MakeVariant("VRRP specifies an election protocol that dynamically assigns responsibility for a virtual router to one of the VRRP routers on a LAN."),
				"protocols":   dbus.MakeVariant([]string{"112"}),
				"destination": dbus.MakeVariant(map[string]string{
					"ipv4": "224.0.0.18",
					"ipv6": "FF02:0:0:0:0:0:0:12",
				}),
			},
			want: Service{
				Name:        "vrrp",
				Short:       "Virtual Router Redundancy Protocol RFC3768 and RFC5798.",
				Description: "VRRP specifies an election protocol that dynamically assigns responsibility for a virtual router to one of the VRRP routers on a LAN.",
				Destination: []Destination{
					{IPFamily: "ipv4", IPAddress: "224.0.0.18"},
					{IPFamily: "ipv6", IPAddress: "FF02:0:0:0:0:0:0:12"},
				},
				Protocols:   []string{"112"},
				Ports:       []Port{},
				SourcePorts: []Port{},
				Includes:    []string{},
			},
		},
		"source ports are accepted": {
			service: "llmnr-client",
			settings: map[string]dbus.Variant{
				"short":        dbus.MakeVariant("LLMNR Client"),
				"description":  dbus.MakeVariant("Link-Local Multicast Name Resolution (LLMNR) allows both IPv4 and IPv6 hosts to perform name resolution for hosts on the same local network. This service allows incoming LLMNR responses. Due to protocol details the service matches by source port and thus allows unsolicited responses."),
				"source_ports": portVariant([2]string{"5355", "udp"}),
			},
			want: Service{
				Name:        "llmnr-client",
				Short:       "LLMNR Client",
				Description: "Link-Local Multicast Name Resolution (LLMNR) allows both IPv4 and IPv6 hosts to perform name resolution for hosts on the same local network. This service allows incoming LLMNR responses. Due to protocol details the service matches by source port and thus allows unsolicited responses.",
				SourcePorts: []Port{{Port: "5355", Protocol: "udp"}},
				Ports:       []Port{},
				Destination: []Destination{},
				Protocols:   []string{},
				Includes:    []string{},
			},
		},
		"includes are accepted": {
			service: "samba",
			settings: map[string]dbus.Variant{
				"short":       dbus.MakeVariant("Samba"),
				"description": dbus.MakeVariant("This option allows you to access and participate in Windows file and printer sharing networks. You need the samba package installed for this option to be useful."),
				"includes":    dbus.MakeVariant([]string{"samba-client"}),
				"ports":       portVariant([2]string{"139", "tcp"}, [2]string{"445", "tcp"}),
			},
			want: Service{
				Name:        "samba",
				Short:       "Samba",
				Description: "This option allows you to access and participate in Windows file and printer sharing networks. You need the samba package installed for this option to be useful.",
				Ports: []Port{
					{Port: "139", Protocol: "tcp"},
					{Port: "445", Protocol: "tcp"},
				},
				Includes:    []string{"samba-client"},
				Destination: []Destination{},
				Protocols:   []string{},
				SourcePorts: []Port{},
			},
		},
		"unknown keys are ignored": {
			service: "ssh",
			settings: map[string]dbus.Variant{
				"short":       dbus.MakeVariant("SSH"),
				"nonexistent": dbus.MakeVariant([]string{"invalid"}),
			},
			want: Service{
				Name:        "ssh",
				Short:       "SSH",
				Ports:       []Port{},
				Destination: []Destination{},
				Protocols:   []string{},
				SourcePorts: []Port{},
				Includes:    []string{},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := serviceFromSettings(test.service, test.settings)
			if err != nil {
				t.Fatalf("serviceFromSettings() returned an error: %v", err)
			}

			if diff := cmp.Diff(test.want, got, cmpopts.SortSlices(func(x, y Destination) bool {
				return x.IPFamily < y.IPFamily
			})); diff != "" {
				t.Errorf("serviceFromSettings() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestServiceFromSettingsRejectsUnexpectedTypes(t *testing.T) {
	tests := map[string]struct {
		service  string
		settings map[string]dbus.Variant
	}{
		"ports as a list of strings are rejected": {
			service: "unexpected-string-list",
			settings: map[string]dbus.Variant{
				"ports": dbus.MakeVariant([]string{"22/tcp"}),
			},
		},
		"source_ports as list of strings are rejected": {
			service: "unexpected-string-list",
			settings: map[string]dbus.Variant{
				"source_ports": dbus.MakeVariant([]string{"5355/udp"}),
			},
		},
		"destination as a list of strings are rejected": {
			service: "unexpected-string-list",
			settings: map[string]dbus.Variant{
				"destination": dbus.MakeVariant([]string{"ipv4"}),
			},
		},
		"includes as a single string is rejected": {
			service: "unexpected-base-type",
			settings: map[string]dbus.Variant{
				"includes": dbus.MakeVariant("test"),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := serviceFromSettings(test.service, test.settings)
			if err == nil {
				t.Fatalf("serviceFromSettings() = %+v, want an error", got)
			}
		})
	}
}

func TestStoreVariantMap(t *testing.T) {
	t.Run("returns the decoded value when the key is present", func(t *testing.T) {
		want := "SSH"
		key := "short"

		m := map[string]dbus.Variant{key: dbus.MakeVariant(want)}

		got, err := storeVariantMap(m, key, "fallback")
		if err != nil {
			t.Fatalf("storeVariantMap() returned an error: %v", err)
		}
		if got != want {
			t.Errorf("storeVariantMap() = %q, want %q", got, want)
		}
	})

	t.Run("returns the fallback when the key is absent", func(t *testing.T) {
		want := "fallback"
		key := "short"

		m := map[string]dbus.Variant{}

		got, err := storeVariantMap(m, key, want)
		if err != nil {
			t.Fatalf("storeVariantMap() returned an error: %v", err)
		}
		if got != want {
			t.Errorf("storeVariantMap() = %q, want %q", got, want)
		}
	})

	t.Run("returns an error when the value has an invalid type", func(t *testing.T) {
		key := "ports"

		m := map[string]dbus.Variant{key: dbus.MakeVariant([]string{"22/tcp"})}

		got, err := storeVariantMap(m, key, []Port{})
		if err == nil {
			t.Fatalf("storeVariantMap() = %+v, want an error", got)
		}
	})
}

type failingObject struct {
	dbus.BusObject
	err error
}

func (f *failingObject) CallWithContext(context.Context, string, dbus.Flags, ...any) *dbus.Call {
	return &dbus.Call{Err: f.err}
}

func TestDBusCallFailuresAreTagged(t *testing.T) {
	internalErr := errors.New("error during communication")
	client := &dbusClient{object: &failingObject{err: internalErr}}

	calls := map[string]func() error{
		"GetDefaultZone": func() error { _, err := client.GetDefaultZone(t.Context()); return err },
		"GetActiveZones": func() error { _, err := client.GetActiveZones(t.Context()); return err },
		"GetServices":    func() error { _, err := client.GetServices(t.Context(), "public"); return err },
		"GetServiceInfo": func() error { _, err := client.GetServiceInfo(t.Context(), "ssh"); return err },
	}

	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			err := call()
			if err == nil {
				t.Fatalf("%s() want an error", name)
			}
			if !errors.Is(err, ErrIPC) {
				t.Errorf("errors.Is(err, ErrIPC) = false for %v", err)
			}
			if !errors.Is(err, internalErr) {
				t.Errorf("error is not wrapped: %v", err)
			}
		})
	}
}
