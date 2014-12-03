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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"psiphon"
	"strings"
)

type ProxyConfig struct {
	httpProxyAddress  string
	httpProxyPort     int
	useHttpProxy      bool
	socksProxyAddress string
	socksProxyPort    int
	useSocksProxy     bool
}

//TODO all of LoadServerEntryConfig.  Modelled after psiphon.LoadConfig

func LoadServerEntryConfig(filename string) (remoteServerList *psiphon.RemoteServerList, err error) {
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, psiphon.ContextError(err)
	}

	err = json.Unmarshal(fileContents, &remoteServerList)
	if err != nil {
		return nil, psiphon.ContextError(err)
	}

	//err = validateRemoteServerlist(?)

	for _, encodedServerEntry := range strings.Split(remoteServerList.Data, "\n") {
		serverEntry, err := psiphon.DecodeServerEntry(encodedServerEntry)
		if err != nil {
			return nil, psiphon.ContextError(err)
		}

		//TODO Remove this
		fmt.Println(serverEntry)
		//TODO Evaluate this, probably not needed
		//     StoreServerEntry puts the serverEntry into a sqlite db.
		err = psiphon.StoreServerEntry(serverEntry, true)
		if err != nil {
			return nil, psiphon.ContextError(err)
		}
	}
	//TODO probably don't need to return remoteServerList
	return remoteServerList, nil
}
