# Electrum tcp

This example shows how to connect to an electrum enabled server using `telnet`. It is assumed that the server is running on the local machine, with the tcp port of
the electrum server listening on `port 7001`.

1. Start the telnet client

```bash
void@Abyssus:[/home/void]> telnet
telnet>
```

2. Create a tcp connection to the server

```bash
telnet> open localhost 7001
Trying ::1...
Connection failed: Connection refused
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
```

3. Verify that the server is repsonding

```bash
{"jsonrpc":"2.0", "method":"server.ping", "id":1}
{"jsonrpc":"2.0","id":1}
```

4. Negotiate the protocol version

```bash
{"jsonrpc":"2.0", "method":"server.version", "params":{"protocol_version":"1.0"}, "id": 2}
{"jsonrpc":"2.0","result":{"server_software_version":"","protocol_version":"1.0"},"id":2}
```

5. Subscribe to an address

```bash
{"jsonrpc":"2.0", "method":"blockchain.address.subscribe", "params":{"address":"015a080a9259b9d4aaa550e2156f49b1a79a64c7ea463d810d4493e8242e6791584fbdac553e6f"}, "id":3}
{"jsonrpc":"2.0","result":{"status":"fb3995dc380e373d3a6eee90695db3e9628150ff498608824ed62bd07b36929f"},"id":3}
```

The client is now subscribed to the address, and will receive updates any time the address changes. Note that the server will close the connection without warning
if we do not send a new request in 10 minutes.