// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.
//
// This inventory script is compatable with all versions of Ansible
// including support for the "_meta" section used by Ansible 1.3. This
// script also supports returning data for a single host will called
// with the --host flag.
//
// This script prints to STDOUT a valid Ansible inventory. For example:
//
//  {
//    "_meta": {
//      "hostvars": {
//        "hostname": {
//          "key": "value"
//        }
//      }
//    },
//    "group_name": {
//      "hosts": ["host.example.com"],
//      "vars": {
//        "key": "value"
//      }
//    }
//  }
//
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	dataDir     = "inventory"
	defaultVars map[string]interface{}
	environment = "production"
	groupDir    = filepath.Join(dataDir, environment, "groups")
	hostDir     = filepath.Join(dataDir, environment, "hosts")
)

type Group struct {
	Hosts []string
	Vars  map[string]interface{}
}

type Host struct {
	Vars map[string]interface{}
}

func main() {
	// Ansible versions prior to 1.3 will call this inventory script once
	// for every host in the group defined by the playbook being executed.
	// In this case we simply output the variables for the host identified
	// by the --host flag, and exit.
	if len(os.Args) == 3 {
		if os.Args[1] == "--host" {
			printHostVarsAndExit()
		}
	}

	inventory := make(map[string]interface{})

	// Ansible supports setting default group vars via a special group
	// named "all". We do something similar here.
	if err := setDefaultVars(); err != nil {
		log.Fatal(err.Error())
	}

	groupPaths, err := ioutil.ReadDir(groupDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	// This is where we start building the bulk of the Ansible inventory
	// described above.  Since group vars are optional, we only add them
	// to the output of the group if non-empty.
	for _, group := range groupPaths {

		// The all group is only used for setting default group vars, it
		// will not be part of the final output.
		if group.Name() == "all.json" {
			continue
		}

		var g Group
		groupMap := make(map[string]interface{})
		groupVars := make(map[string]interface{})
		groupName := trimExt(group.Name(), ".json")

		if defaultVars != nil {
			for k, v := range defaultVars {
				groupVars[k] = v
			}
		}

		err = unmarshalFromFile(filepath.Join(groupDir, group.Name()), &g)
		if err != nil {
			log.Fatal(err.Error())
		}

		groupMap["hosts"] = g.Hosts

		if g.Vars != nil {
			for k, v := range g.Vars {
				groupVars[k] = v
			}
		}
		if len(groupVars) > 0 {
			groupMap["vars"] = groupVars
		}
		inventory[groupName] = groupMap
	}

	// Starting with Ansible 1.3, we can output a top level element named
	// "_meta".  When "_meta" contains a value for "hostvars" we can return
	// all host vars upfront and avoid additional calls to this inventory
	// script for each host.
	//
	// The "_meta" section has the following output:
	//
	//  {
	//    "_meta": {
	//      "hostvars": {
	//        "host.example.com": {"key": "value"}
	//      }
	//    }
	//  }
	//
	meta := make(map[string]interface{})
	hostvars := make(map[string]interface{})
	meta["hostvars"] = hostvars

	hosts, err := ioutil.ReadDir(hostDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, host := range hosts {
		hostVars, err := getHostVars(filepath.Join(hostDir, host.Name()))
		if err != nil {
			log.Fatal(err.Error())
		}

		if len(hostVars) == 0 {
			continue
		}
		hostvars[trimExt(host.Name(), ".json")] = hostVars
	}
	inventory["_meta"] = meta

	// Format the final inventory output and print to STDOUT.
	output, err := json.MarshalIndent(inventory, "", "  ")
	fmt.Printf(string(output))
}

func printHostVarsAndExit() {
	fpath := filepath.Join(hostDir, os.Args[2]+".json")
	if isFileExist(fpath) {
		hostVars, err := getHostVars(fpath)
		if err != nil {
			log.Fatal(err.Error())
		}
		data, err := json.MarshalIndent(hostVars, "", "  ")
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("%s\n", string(data))
		os.Exit(0)
	} else {
		fmt.Print("{}\n")
		os.Exit(0)
	}
}

func setDefaultVars() error {
	var g Group
	filePath := filepath.Join(groupDir, "all.json")
	if isFileExist(filePath) {
		err := unmarshalFromFile(filePath, &g)
		if err != nil {
			return err
		}
		defaultVars = g.Vars
	}
	return nil
}

func getHostVars(fpath string) (map[string]interface{}, error) {
	var h Host
	if err := unmarshalFromFile(fpath, &h); err != nil {
		return nil, err
	}
	return h.Vars, nil
}

func unmarshalFromFile(fpath string, dest interface{}) error {
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, dest); err != nil {
		return err
	}
	return nil
}

func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

func trimExt(s, ext string) string {
	if filepath.Ext(s) != ext {
		return s
	}
	return s[0 : len(s)-len(ext)]
}
