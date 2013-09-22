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

	// Ansible expects JSON on stdout representing the inventory, which
	// should have the following format:
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
	inventory := make(map[string]interface{})

	groupPaths, err := ioutil.ReadDir(groupDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Ansible supports setting default group vars for all groups via
	// the all group. We do something similar here.
	if err := setDefaultVars(); err != nil {
		log.Fatal(err.Error())
	}

	// Start building Group section
	for _, group := range groupPaths {
		if group.Name() == "all.json" {
			continue
		}

		var g Group
		groupMap := make(map[string]interface{})
		groupVars := make(map[string]interface{})
		name := trimExt(group.Name(), ".json")

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

		inventory[name] = groupMap
	}

	// Add _meta section
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

	// Format inventory and print to stdout.
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
