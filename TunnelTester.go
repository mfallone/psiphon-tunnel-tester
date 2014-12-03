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
	"fmt"
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

	flag.Parse()

	if configFilename == "" {
		log.Fatalf("configuration file is required")
	}

	config, err := psiphon.LoadConfig(configFilename)
	if err != nil {
		log.Fatalf("error loading configuration file: %s", err)
	}

	//TODO find a more useful place for this.
	pendingConns := new(psiphon.Conns)
	localHttpProxyAddress := "127.0.0.1"
	localHttpProxyPort := config.LocalHttpProxyPort
	useProxy := false

	site := "http://vl7.net/ip"
	untunneledCheck, err := GetWebResource(site, useProxy, localHttpProxyAddress, localHttpProxyPort)
	if err != nil {
		fmt.Println("Error getting resource: %s", err)
	}
	fmt.Println("Untunneled IP: ", string(untunneledCheck))

	// Check for a server entry string at the cli.  It supercedes the other lists
	if serverEntryString != "" {
		decodedServerString, err := psiphon.DecodeServerEntry(serverEntryString)
		if err != nil {
			log.Fatalf("Invalid server entry, %s", err)
		}
		// TODO: get rid of this
		fmt.Println("Decoded Server Entry: ", decodedServerString)

		tunnel, err := psiphon.EstablishTunnel(config, pendingConns, decodedServerString)
		if err != nil {
			log.Fatalf("Could not establish tunnel: %s", err)
		}

		httpProxy, err := psiphon.NewHttpProxy(config, tunnel)
		if err != nil {
			log.Fatalf("error initializing local HTTP proxy: %s", err)
		}
		defer httpProxy.Close()
		useProxy = true

		tunneledCheck, err := GetWebResource(site, useProxy, localHttpProxyAddress, localHttpProxyPort)
		if err != nil {
			fmt.Println("Error getting resource: ", err)
		}
		fmt.Println("Tunneled IP Check: ", string(tunneledCheck))
	} else if serverEntryFilename != "" {
		log.Println("Attempting to use server entry from file")
		serverEntryConfig, err := LoadServerEntryConfig(serverEntryFilename)
		if err != nil {
			log.Fatalf("error loading server entry file: %s", err)
		}
		// TODO: Remove this
		log.Println("Server Entry: %s", serverEntryConfig)

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

	// test print
	fmt.Println(config.PropagationChannelId)

	//serverEntry is an encoded string and needs to be decoded
}
