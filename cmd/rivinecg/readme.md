Rivinecg Usage
==========

`rivinecg` is the command line interface to Rivine Blockchain Generator.

Generate tasks
------------

Generate:
* `rivinecg generate config [-o/--output]` generate blockchain config file
* `rivinecg generate blockchain` generate blockchain from a config file
* `rivinecg generate seed [-n]` generate a seed and one or multiple addresses

Full Descriptions
-----------------

#### Generate tasks

* `rivinecg generate config [-o/--output]` generates a default blockchain config file.
with the `-o` flag you can provide a path where the file will be stored. Encoding is based on the file extension that you provide,
can be `yaml` or `json`. Default the file will be named `blockchaincfg.yaml` and will be stored in the directory from where you call the command.

* `rivinecg generate blockchain [-c/--config] [-o/--output]` generates a fully working blockchain code directory based on a config file.
the argument `-c` is required and needs to be a path where a config file is stored.
By default the location of your config file is used, another output path can be defined using the -`o` flag.

* `rivinecg generate seed [-n]` generates a seed and matching addresses.
with the `-n` flag you can provide how many addresses should be generated with this seed. These addresses can be used to provide as addresses in a config file.


#### General commands

* `rivinecg version` displays the version string of rivinecg.
