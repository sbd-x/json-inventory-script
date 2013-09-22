package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var dataDir = "inventory"
var environment = "production"

type Config struct {
	GroupDir string
	HostDir  string
}

type Group struct {
	name  string
	Hosts []string
	Vars  map[string]interface{}
}
type Host struct {
	name string
	Vars map[string]interface{}
}

func main() {
	var c Config
	c.GroupDir = filepath.Join(dataDir, environment, "groups")
	c.HostDir = filepath.Join(dataDir, environment, "hosts")

	// Check if we are being called for a specific host.
	if len(os.Args) == 3 {
		if os.Args[1] == "--host" {
			fpath := filepath.Join(c.HostDir, os.Args[2]+".json")
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
	}

	// Inventory
	inventory := make(map[string]interface{})

	groups, err := ioutil.ReadDir(c.GroupDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	var defaultGroup Group
	data, err := ioutil.ReadFile(filepath.Join(c.GroupDir, "all.json"))
	if err != nil {
		log.Fatal(err.Error())
	}

	err = json.Unmarshal(data, &defaultGroup)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, group := range groups {
		if group.Name() == "all.json" {
			continue
		}

		destMap := make(map[string]interface{})
		for k, v := range defaultGroup.Vars {
			destMap[k] = v
		}

		f, err := ioutil.ReadFile(filepath.Join(c.GroupDir, group.Name()))
		if err != nil {
			log.Fatal(err.Error())
		}

		var g Group
		err = json.Unmarshal(f, &g)
		if err != nil {
			log.Fatal(err.Error())
		}

		gMap := make(map[string]interface{})
		gMap["hosts"] = g.Hosts

		if g.Vars != nil {
			for k, v := range g.Vars {
				destMap[k] = v
			}
		}

		if len(destMap) > 0 {
			gMap["vars"] = destMap
		}
		g.name = trimExt(group.Name())

		inventory[g.name] = gMap
	}

	meta := make(map[string]interface{})
	hostvars := make(map[string]interface{})
	meta["hostvars"] = hostvars

	hosts, err := ioutil.ReadDir(c.HostDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, host := range hosts {
		hostVars, err := getHostVars(filepath.Join(c.HostDir, host.Name()))
		if err != nil {
			log.Fatal(err.Error())
		}

		if len(hostVars) == 0 {
			continue
		}
		hostvars[trimExt(host.Name())] = hostVars
	}

	inventory["_meta"] = meta
	output, err := json.MarshalIndent(inventory, "", "  ")
	fmt.Printf(string(output))
}

func getHostVars(fpath string) (map[string]interface{}, error) {
	var h Host
	m := make(map[string]interface{})

	f, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(f, &h); err != nil {
		return m, err
	}
	if h.Vars == nil {
		return m, nil
	} else {
		return h.Vars, nil
	}
}

func isFileExist(fpath string) bool {
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func trimExt(s string) string {
	// check if we have the extension, if not return
	// the input string.
	return s[0 : len(s)-len(".json")]
}
