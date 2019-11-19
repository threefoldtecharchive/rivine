# Caddy plugin

## Motivation

The goal of this plugin is to enable a high availability setup for explorers.
Given that we are a blockchain application, meaning every instance might have a
different view of what it considers to be the active chain and therefore "the truth",
we need to overcome some extra obstacles. Most importantly, the definition of a
"healthy node" is more strict. We can't just say any node which responds is healthy,
for instance, a node which has been down for a really long time will need some additional
time to sync up before it should serve requests again. Although is is perfectly fine
for this node to reply, it does not have the latest blocks, or the chain might have forked
in the meantime. The easy setup, which runs every explorer on it's own DNS, puts the burden
of defining healthy nodes on the clients. Rather than picking any node which they can
get a reply from, They need to connect to *all* nodes which reply, get their block heights,
count how many times a certain hight occurs, and only mark instances on that height as healthy.
Failure to do this, (and randomly sending requests to different nodes), can lead to
an inconsistent view of the data by the client, where (recent) transactions might
appear and disappear, for example when a new instance is syncing. On the other hand,
a regular HA approach where the client connects to a single DNS, which is set up
with multiple instances with failover is not ideal either, since such setups
generally don't allow this custom type of healthcheck, but just care whether
the instance is responding or not. To this end this caddy plugin has been created,
which is a modification of the standard caddy _proxy_ plugin. Most importantly,
the healtcheck code has been updated to work as described above. This means that, an
application needs to connect to just one server, the server running the caddy
which has this plugin. Note that this means that this caddy is now a single point of
failure. Traditional HA still needs to be applied now, but this custom plugin
does allow all the healthcheck logic to be offloaded to the server side.

## Building

To create a caddy with this module enabled, create a directory and add the following
_main.go_ file:

```
package main

import (
	"fmt"

	"github.com/caddyserver/caddy"
	"github.com/caddyserver/caddy/caddy/caddymain"
	"github.com/caddyserver/caddy/caddyhttp/httpserver"

	_ "github.com/threefoldtech/rivine/caddy-plugin"
)

func main() {
	httpserver.RegisterDevDirective("reb", "proxy")
	caddymain.Run()
}
```
Now run the following commands:

```
go mod init caddy
go get github.com/caddyserver/caddy
go install
```

You will now have a caddy server with this plugin enabled in your `$GOPATH`.

Notice that we registered our plugin in the main function here. According to the
official caddy docs, it is recommended to manually register the plugin, by inserting
the name in the plugin list. This approach still works though, especially if only 1
plugin is used, but it does print a warning at startup.

You can now use the `reb` directive like you would normally use the `proxy` directive.
