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
	"io/ioutil"
	"log"

	"github.com/Psiphon-Labs/psiphon-tunnel-core/psiphon"
)

func main() {
	var configFilename string
	flag.StringVar(&configFilename, "config", "psiphon.config", "configuration file")

	var serverEntryFilename string
	flag.StringVar(&serverEntryFilename, "serverList", "server_list.config", "server entry file")

	var encodedServerEntry string
	flag.StringVar(&encodedServerEntry, "serverEntry", "", "encoded server entry")

	var tasksConfigFilename string
	flag.StringVar(&tasksConfigFilename, "tasksConfig", "tasks.config", "Tests config file")

	flag.Parse()

	if configFilename == "" {
		log.Fatalf("configuration file is required")
	}

	configFileContents, err := ioutil.ReadFile(configFilename)
	if err != nil {
		log.Fatalf("error loading configuration file: %s", err)
	}

	config, err := psiphon.LoadConfig(configFileContents)
	if err != nil {
		log.Fatalf("error loading configuration file: %s", err)
	}

	err = psiphon.InitDataStore(config)
	if err != nil {
		log.Fatalf("error initializing datastore: %s", err)
	}

	if tasksConfigFilename == "" {
		log.Fatalln("No tests file specified.  Set ARG --tests-config")
	}
	tasksConfig, err := LoadTasksConfig(tasksConfigFilename)
	if err != nil {
		log.Fatalf("Could not load tasks config file: %s", err)
	}

	serverEntry := new(psiphon.ServerEntry)

	// Check for a server entry string at the cli.  It supercedes the other lists
	if encodedServerEntry != "" {
		serverEntry, err = psiphon.DecodeServerEntry(encodedServerEntry)
		if err != nil {
			log.Fatalf("Invalid server entry, %s", err)
		}

		if psiphon.ValidateServerEntry(serverEntry) != nil {
			log.Fatalf("Could not validate server entry")
		}

		//SetupTasks(config, serverEntry, *tasksConfig)

	} else if serverEntryFilename != "" {
		log.Println("Attempting to use server entry from file")
		serverEntryConfig, err := LoadServerEntryConfig(serverEntryFilename)
		if err != nil {
			log.Fatalf("error loading server entry file: %s", err)
		}

		_, err = psiphon.DecodeServerEntry(serverEntryConfig.Data)

	} else if config.TargetServerEntry != "" {
		serverEntry, err = psiphon.DecodeServerEntry(config.TargetServerEntry)
		if err != nil {
			log.Fatalf("Could not load TargetServerEntry from config: Error: %v", err)
		}
	}

	if config != nil && serverEntry != nil && tasksConfig != nil {
		SetupTasks(config, serverEntry, *tasksConfig)
	}
}
