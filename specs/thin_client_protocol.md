# thin client protocol

## Problem
Thin clients communicate with  explorers/gateways over https/REST.
There is no subscribe mechanism to allow push notifications so polling is required. 
On top of this, the exposed services and json data formats are all custom to Rivine.

### Why to Change Something That Works?
Let's answer with Slush's stratum reasons:

**HTTP: Communication is Driven by clients**
... However  servers know much better when clients need new blocks or utxo's . HTTP was designed for web site browsing where clients ask servers for specific content. Wallets  of a blockchain are  different - the server knows very well what clients need and can control the communication in a more efficient way. Let’s swap roles and leave orchestration to the server.

**Long Polling: An Anti-Pattern**
Long polling uses separate connections  to the  servers, which leads to various issues on server side, like load balancing of connections between more backends. Load balancing using IP hashes or sticky HTTP sessions are just  workarounds for keeping all that stuff working.

Another problem consists of packet storms, coming from clients trying to reconnect to the server after long polling broadcasts. Sometimes it's hard to distinguish valid long polling reconnections from DDoS attacks. All this makes the architecture more complicated and harder to maintain, which is reflected in a less reliable  service and has a real impact on clients.

## Proposed solution

 Instead of reinventing the wheel, The [electrum protocol](https://electrumx.readthedocs.io/en/latest/protocol.html) can be implemented.

 Issues: https://github.com/threefoldtech/rivine/issues/385

