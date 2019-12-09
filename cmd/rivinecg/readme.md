# Rivine based blockchain generator

`rivinecg` is a tool for generating Rivine based Blockchains.

## Summary of commands

Generate:

* `rivinecg generate seed [-n]` generates a seed and one or multiple addresses
* `rivinecg generate config [-o/--output]` generates blockchain config file
* `rivinecg generate blockchain` generates blockchain from a config file

## Commands descriptions

### Generate tasks

* `rivinecg generate seed [-n]` generates a seed and matching addresses.
with the `-n` flag you can provide how many addresses should be generated with this seed. These addresses can be used to provide as addresses in a config file.

* `rivinecg generate config [-o/--output]` generates a default blockchain config file.
with the `-o` flag you can provide a path where the file will be stored. Encoding is based on the file extension that you provide,
can be `yaml` or `json`. Default the file will be named `blockchaincfg.yaml` and will be stored in the directory from where you call the command.
Be sure to modify the addresses in the generated configfile.

* `rivinecg generate blockchain [-c/--config] [-o/--output]` generates a fully working blockchain code directory based on a config file.
the argument `-c` is required and needs to be a path where a config file is stored.
By default the location of your config file is used, another output path can be defined using the -`o` argument.

### Show help

`rivinecg --help`
