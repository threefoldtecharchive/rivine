# Protocol methods

## Server.ping

Ping the server to make sure it is responding, and to keep the session alive.

### Params

- None

Example Request:

```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "server.ping"
}
```

### Result

- None

```json
{
    "jsonrpc": "2.0",
    "id": 1
}
```

-----

## server.version

Negotiate protocol version to use and optionally identify the client to the server

### Params

- client_name (optional)

A string identifying the client software. It is not directly used by the server for any functional purpose, but could be used for debug purposes.

- protocol_version

An array `[protocol_min, protocol_max]`, each of which is a string. If `protocol_min` and `protocol_max` are the same, they can be passed
as a single string instead. If there is no version that both the client and the server understand, the connection will be closed by the server.
The protocol version is denoted as `"a.b[.c]"`, `a` being the major number, `b` being the minor number, and `c` being the revision. If `c` is 0, it can be ommited.

Currently only version `"1.0"` is supported.

Example Request:

```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "server.version",
    "params": {
        "client_name": "my_electrum_client_v1",
        "protocol_version": "1.0"
    }
}
```

### Result

- server_software_version

The version of the server running. This is the build version of the server. This allows to exactly identify the state of the server code base,
usefull should you report a bug or a feature request.

- protocol_version

The version of the protocol which will be used when talking to the server.

Example Response:

```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "result": {
        "server_software_version": "v1.0.8",
        "protocol_version": "1.0"
    }
}
```

-----

## blockchain.address.subscribe

### Params

- address

The string representation of an address.

Example Request:

```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "blockchain.address.subscribe",
    "params": {
        "address": "015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"
    }
}
```

### Result

- status

The [status](./Status.md#address-status) of the address.

Example Response:

```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": {
        "status": "fb3995dc380e373d3a6eee90695db3e9628150ff498608824ed62bd07b36929f"
    }
}
```

### Notifications

As a subscribtion, a notification will be send to the client every time the address [status](./Status.md#address-status) changes

#### Params

- address

  The address which changed [status](./Status.md#address-status)

- status
  
  The new [status](./Status.md#address-status) of the address

Example Notification:

```json
{
    "jsonrpc": "2.0",
    "method": "blockchain.address.subscribe",
    "params": {
        "address": "015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f",
        "status": "fb3995dc380e373d3a6eee90695db3e9628150ff498608824ed62bd07b36929f"
    }
}
```
