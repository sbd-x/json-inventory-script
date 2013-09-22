# Ansible Inventory Script

## Inventory Layout

This inventory script expects a specific file layout representing the Ansible inventory. All files and directory live under the `datadir`. Host and Groups are group by environment. All groups and hosts data files are named after the group or host and must be valid JSON.

```
├── inventory
│   └── production
│       ├── groups
│       │   ├── 5points.json
│       │   ├── all.json
│       │   ├── atlanta.json
│       │   ├── databases.json
│       │   ├── marietta.json
│       │   └── webservers.json
│       └── hosts
│           ├── llama.example.com.json
│           └── moocow.example.com.json
```

### Group files

Groups can define the following keys:

 * hosts
 * vars

### Host files

Hosts can define the following keys:

 * vars

## Setting default values for all groups

Default values are defined in the `all` group's data file. In the case of the production environment:

```
$datadir/production/groups/all.json
```
