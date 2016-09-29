Rivine 0.0.1
============

[![Build Status](https://travis-ci.org/rivine/rivine.svg?branch=master)](https://travis-ci.org/rivine/rivine)
[![GoDoc](https://godoc.org/github.com/rivine/rivine?status.svg)](https://godoc.org/github.com/rivine/rivine)
[![Go Report Card](https://goreportcard.com/badge/github.com/rivine/rivine)](https://goreportcard.com/report/github.com/rivine/rivine)

Usage
-----

This release comes with 2 binaries, rivined and rivinec. rivined is a background
service, or "daemon," that runs the Sia protocol, and rivinec is a client that is
used to interact with rivined. rivined exposes an HTTP API on 'localhost:23110' which
can be used to interact with the daemon. Documentation on the API can be found in doc/API.md.

rivined and rivinec are run via command prompt. On Windows, you can just double-
click rivined.exe if you don't need to specify any command-line arguments.
Otherwise, navigate to the rivine folder and click File->Open command prompt.
Then, start the rivined service by entering `rivined` and pressing Enter. The
command prompt may appear to freeze; this means rivined is waiting for requests.
Windows users may see a warning from the Windows Firewall; be sure to check
both boxes ("Private networks" and "Public networks") and click "Allow
access." You can now run `rivinec` in a separate command prompt to interact with
rivined. From here, you can send money, mine blocks, upload and download
files, and advertise yourself as a host.

Building From Source
--------------------

To build from source, [Go 1.6 must be installed](https://golang.org/doc/install)
on the system. Then simply use `go get`:

```
go get -u github.com/rivine/rivine/...
```

This will download the Rivine repo to your `$GOPATH/src` folder, and install the
`rivined` and `rivinec` binaries in your `$GOPATH/bin` folder.

To stay up-to-date, run the previous `go get` command again. Alternatively, you
can use the Dockerfile provided in this repo. Run `docker build -t rivine .`
to build and `docker run --name rivine rivine` to start the daemon.
Running the client can be done with `docker run -it rivine rivinec`.
Add client commands just like you would calling rivinec normally (like `docker run -it rivine rivinec wallet transactions`).


Troubleshooting
---------------

- I can't connect to more than 8 peers.

  Once Sia has connected to 8 peers, it will stop trying to form new
  connections, but it will still accept incoming connection requests (up to 128
  total peers). However, if you are behind a firewall, you will not be able to
  accept incoming connections. You must configure your firewall to allow Rivine
  connections by forwarding your ports. By default, Rivine communicates on port
  23112. The specific instructions for forwarding a port vary by
  router. For more information, consult [this guide](http://portfoward.com).

  Rivine currently has support for UPnP. While not all routers support UPnP, a
  majority of users should have their ports automatically forwarded by UPnP.


Version Information
-------------------


Version History
---------------
