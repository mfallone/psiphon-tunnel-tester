/*
 * Copyright (c) 2014, Psiphon Inc.
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

/*
 * This app will test Psiphon tunnels.
 * It will consume a configuration file containing a single Psiphon server
 * create a tunnel based on the configuration and test connectivity through
 * said tunnel.
 * There can be a variety of connectivity tests but for now a simple http request
 * to find verified that the traffic is tunneled should suffice.
 */

package main

import (
	"flag"
	"log"

	psiphon "github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
)

func main() {
	var configFilename string
	flag.StringVar(&configFilename, "config", "psiphon.config", "configuration file")

	var serverEntryFilename string
	flag.StringVar(&serverEntryFilename, "server-list", "server_list.config", "server entry file")

	var serverEntryString string
	flag.StringVar(&serverEntryString, "server-entry", "", "encoded server entry")

	var tasksConfigFilename string
	flag.StringVar(&tasksConfigFilename, "tasks-config", "tasks.config", "Tests config file")

	flag.Parse()

	if configFilename == "" {
		log.Fatalf("configuration file is required")
	}

	config, err := psiphon.LoadConfig(configFilename)
	if err != nil {
		log.Fatalf("error loading configuration file: %s", err)
	}

	if tasksConfigFilename == "" {
		log.Fatalln("No tests file specified.  Set ARG --tests-config")
	}
	tasksConfig, err := LoadTasksConfig(tasksConfigFilename)
	if err != nil {
		log.Fatalf("Could not load tasks config file: %s", err)
	}

	// Check for a server entry string at the cli.  It supercedes the other lists
	if serverEntryString != "" {
		decodedServerString, err := psiphon.DecodeServerEntry(serverEntryString)
		if err != nil {
			log.Fatalf("Invalid server entry, %s", err)
		}
		//Run Tests
		_, err = RunTests(config, decodedServerString, tasksConfig)
		if err != nil {
			log.Fatalf("Could not run tunnel tests: ", err)
		}

	} else if serverEntryFilename != "" {
		log.Println("Attempting to use server entry from file")
		serverEntryConfig, err := LoadServerEntryConfig(serverEntryFilename)
		if err != nil {
			log.Fatalf("error loading server entry file: %s", err)
		}

		decodedServerString, err := psiphon.DecodeServerEntry(serverEntryConfig.Data)
		// TODO: Remove this
		log.Printf("Server Entry: %s", decodedServerString)

	} else if serverEntryFilename == "" { // Check for server entry file name, if not found try the remote server list
		log.Printf("No server entry file provided, trying remote server list")
		if config.RemoteServerListUrl == "" {
			log.Fatalf("No remote server list found")
		} else {
			// TODO load remote server list
			/*
				err := psiphon.FetchRemoteServerList(config, pendingConns)
				if err != nil {
					log.Fatalf("failed to fetch remote server list: %s", err)
				}
			*/
		}
	}
}
