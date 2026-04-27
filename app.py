# Copyright 2026 SUSE LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
from dataclasses import dataclass

import dbus
from mcp.server.fastmcp import FastMCP


@dataclass
class Port:
    port: str
    protocol: str


@dataclass
class Destination:
    ip_family: str
    ip_addess: str


@dataclass
class Service:
    name: str
    short: str
    description: str
    ports: list[Port]
    destination: list[Destination]
    protocols: list[str]
    source_ports: list[Port]
    includes: list[str]


def fw_get_active_zones(fwd_proxy) -> list[str]:
    zone_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1.zone")
    return [str(zone) for zone in zone_iface.getActiveZones()]


def fw_get_services(fwd_proxy, zone) -> list[str]:
    zone_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1.zone")
    return [str(service) for service in zone_iface.getServices(zone)]


def fw_get_default_zone(fwd_iface) -> str:
    return str(fwd_iface.getDefaultZone())


def fw_get_service_settings(fwd_iface, service_name) -> Service:
    service_obj = fwd_iface.getServiceSettings2(service_name)
    service = Service(
        str(service_name),
        str(service_obj.get("short", "")),
        str(service_obj.get("description", "")),
        [
            Port(str(port), str(protocol))
            for [port, protocol] in service_obj.get("ports", [])
        ],
        [
            Destination(str(family), str(addr))
            for family, addr in service_obj.get("destination", {}).items()
        ],
        [str(protocol) for protocol in service_obj.get("protocols", [])],
        [
            Port(str(port), str(protocol))
            for [port, protocol] in service_obj.get("source_ports", [])
        ],
        [str(include) for include in service_obj.get("includes", [])],
    )
    return service


mcp = FastMCP("firewalld")
bus = dbus.SystemBus()


@mcp.tool()
def get_default_zone() -> str:
    """Get the name of the default firewall zone"""
    fwd_proxy = bus.get_object(
        "org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1"
    )
    fwd_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1")
    return fw_get_default_zone(fwd_iface)


@mcp.tool()
def is_default_zone(zone: str) -> bool:
    """Check whether the specified firewall zone is the default zone

    Args:
        zone: Name of the firewall zone
    """
    fwd_proxy = bus.get_object(
        "org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1"
    )
    fwd_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1")
    return fw_get_default_zone(fwd_iface) == zone


@mcp.tool()
def get_active_zones() -> list[str]:
    """Get the names of all currently active firewall zones"""
    fwd_proxy = bus.get_object(
        "org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1"
    )
    return fw_get_active_zones(fwd_proxy)


@mcp.tool()
def get_services_for_zone(zone: str) -> list[str]:
    """Get all allowed services of a firewall zone

    Args:
        zone: Name of the firewall zone
    """
    fwd_proxy = bus.get_object(
        "org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1"
    )
    fwd_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1")
    return fw_get_services(fwd_iface, zone)


@mcp.tool()
def get_service_info(service: str) -> Service:
    """Get information on the specified service.
    Information include, the service's name, description, ports, destination addresses,
    protocols, source_ports.  When a service contains items in the includes field, you
    need to request data for each listed include.

    Args:
        service: Name of the service
    """
    fwd_proxy = bus.get_object(
        "org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1"
    )
    fwd_iface = dbus.Interface(fwd_proxy, "org.fedoraproject.FirewallD1")
    return fw_get_service_settings(fwd_iface, service)


if __name__ == "__main__":
    mcp.run(transport="stdio")
