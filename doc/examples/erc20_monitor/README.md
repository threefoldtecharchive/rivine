# erc20 monitor

This is a demo application showing the erc20 (T)TFT to regular (T)TFT conversion. Right now the page
is only able to show the balance of an address.

## Contract

A test contract has been deployed on the [Rinkeby testnet](https://rinkeby.etherscan.io/). The contract source can
be found in [testcontract.sol](testcontract.sol).

## Running

If you want to run this example yourself, simply build the example (`go build`), and run the produced binary
in this directory. You will require an [etherscan](https://etherscan.io) API key. You also need to specify the
listening address for the server.

The following example runs the server on `port 8888` on every interface on the local machine, with an `API Key`
set as an environment variable.

```bash
./erc20_monitor $APIKEY :8888
```

## Development

The webpage uses [vue.js](https://vuejs.org). Local development requires `vue cli` for some necessary tools.
Further explanation on the web development process can be found on the official vue docs.