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
var groupDir = filepath.Join(dataDir, environment, "groups")
var hostDir = filepath.Join(dataDir, environment, "hosts")

type Inventory struct {
	Name   string
	Groups []Group
	Hosts  []Host
	Meta   map[string]interface{}
}

type Group struct {
	name  string
	Hosts []string               `json:"hosts"`
	Vars  map[string]interface{} `json:"vars"`
}

type Host struct {
	name string
	Vars map[string]interface{}
}

func main() {
	if len(os.Args) == 3 {
		if os.Args[1] == "--host" {
			_, err := os.Stat(filepath.Join(hostDir, os.Args[2] + ".json"))
			if os.IsNotExist(err) {
				fmt.Print("{}\n")
				os.Exit(0)
			}

			hostVars, err := getHostVars(os.Args[2] + ".json")
			if err != nil {
				log.Fatal(err.Error())
			}

			data, err := json.MarshalIndent(hostVars, "", "  ")
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Printf("%s\n", string(data))

			os.Exit(0)
		}
	}

	m := make(map[string]interface{})

	groups, err := ioutil.ReadDir(groupDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	var defaultGroup Group
	data, err := ioutil.ReadFile(filepath.Join(groupDir, "all.json"))
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

		f, err := ioutil.ReadFile(filepath.Join(groupDir, group.Name()))
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

		m[g.name] = gMap
	}

	meta := make(map[string]interface{})
	hostvars := make(map[string]interface{})
	meta["hostvars"] = hostvars

	hosts, err := ioutil.ReadDir(hostDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, host := range hosts {
		hostVars, err := getHostVars(host.Name())
		if err != nil {
			log.Fatal(err.Error())
		}

		if len(hostVars) == 0 {
			continue
		}
		hostvars[trimExt(host.Name())] = hostVars
	}

	m["_meta"] = meta
	output, err := json.MarshalIndent(m, "", "  ")
	fmt.Printf(string(output))
}

func getHostVars(name string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	f, err := ioutil.ReadFile(filepath.Join(hostDir, name))
	if err != nil {
		return nil, err
	}

	var h Host
	if err = json.Unmarshal(f, &h); err != nil {
		return m, err
	}

	if h.Vars == nil {
		return m, nil
	} else {
		return h.Vars, nil
	}
}

func trimExt(s string) string {
	return s[0 : len(s)-len(".json")]
}
