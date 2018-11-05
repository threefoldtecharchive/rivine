# Address status

The address status is a `sha256` hash of the information regarding the address. The address calculation is implemented as described [in the electrum docs](https://electrumx.readthedocs.io/en/latest/protocol-basics.html#status).
Although it is not possible to construct the state of an address from this status string, a client could choose to save the state of an address localy, using this status string as the key. This would allow the client to fetch
the exact state from local storage, rather then calling over the network again, saving bandwith.