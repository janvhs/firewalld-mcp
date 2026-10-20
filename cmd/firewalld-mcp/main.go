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
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/godbus/dbus/v5"
	"github.com/janvhs/firewalld-mcp/internal/firewalld"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const shaLen = 7

var (
	// MCPName is the name this server announces to MCP clients. It can be set
	// via ldflags when building.
	MCPName = "firewalld"

	// Version contains the application version number. It's set via ldflags
	// when building.
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags when building.
	CommitSHA = ""
)

var flagVersion bool

func init() {
	flag.BoolVar(&flagVersion, "version", false, "Print version and exit")

	if Version == "" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
			Version = info.Main.Version
		} else {
			Version = "unknown (built from source)"
		}
	}
}

func makeUsage(out io.Writer, name string) func() {
	return func() {
		fmt.Fprintf(out, "Usage: %s [options...]\n\nOptions:\n", name)
		flag.PrintDefaults()
	}
}

func main() {
	out := flag.CommandLine.Output()
	name := flag.CommandLine.Name()

	flag.Usage = makeUsage(out, name)
	flag.Parse()

	if flagVersion {
		versionString := fmt.Sprintf("%s %s", name, Version)
		if len(CommitSHA) >= shaLen {
			versionString += " (" + CommitSHA[:shaLen] + ")"
		}

		fmt.Println(versionString)
		return
	}

	if flag.NArg() > 0 {
		fmt.Fprintf(out, "argument provided but not expected: %q\n", flag.Arg(0))
		flag.Usage()
		os.Exit(2)
	}

	if err := mainE(); err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
}

func mainE() error {
	ctx := context.Background()

	bus, err := dbus.ConnectSystemBus()
	if err != nil {
		return fmt.Errorf("connect system bus: %w", err)
	}
	defer func() { _ = bus.Close() }()

	server := newMCPServer(firewalld.NewDBusClient(bus))

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return err
	}

	return nil
}
