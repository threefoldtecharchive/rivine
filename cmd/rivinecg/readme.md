Rivinecg Usage
==========

`rivinecg` is the command line interface to Rivine Blockchain Generator. It comes as a part of the command line
package, and can be run as `./rivinecg` from the same folder, or just by
calling `rivinecg` if you move the binary into your path.

Generate tasks
------------

Generate:
* `rivinecg generate config [-p]` generate blockchain config file
* `rivinecg generate generate-blockchain` generate blockchain from a config file
* `rivinecg generate seed [-n]` generate a seed and one or multiple addresses

Full Descriptions
-----------------

#### Generate tasks

* `rivinecg generate config [-p]` generates a default blockchain config file.
with the `-p` flag you can provide a path where the file will be stored. Encoding is based on the file extension that you provide,
can be `yaml` or `json`. Default the file will be named `blockchaincfg.yaml` and will be stored in the directory from where you call the command.

* `rivinecg generate blockchain [config-file]` generates a fully working blockchain code directory based on a config file.
the argument `config-file` is required and needs to be a path where a config file is stored. The blockchain code directory will be stored in your `GOPATH`.
In the config file there are 2 variables that decide where the blockchain directory will be stored in your `GOPATH`, `owner` and `name`.

Example:

```
blockchain:
  name: foodir
  owner: JohnDoe
```

With these parameters in the config file, the blockchain code will be generated in: `GOPATH/src/github.com/JohnDoe/foodir`.
Owner usualy is a `github name` and name is usualy a `github repository`.

* `rivinecg generate seed [-n]` generates a seed and matching addresses.
with the `-n` flag you can provide how many addresses should be generated with this seed. These addresses can be used to provide as addresses in a config file.


#### General commands

* `rivinecg version` displays the version string of rivinecg.
