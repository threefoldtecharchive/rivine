Daemon API
===========

This document contains detailed descriptions of the daemon's API routes. For an
overview of the daemon's API routes, see [API.md#daemon](/doc/API.md#daemon).
For an overview of all API routes, see [API.md](/doc/API.md)

There may be functional API calls which are not documented. These are not
guaranteed to be supported beyond the current release, and should not be used
in production.

Overview
--------

The daemon is responsible for starting and stopping the modules which make up
the rest of Rivine. It also provides endpoints for viewing build constants.

Index
-----

| Route                                     | HTTP verb |
| ----------------------------------------- | --------- |
| [/daemon/constants](#daemonconstants-get) | GET       |
| [/daemon/version](#daemonversion-get)     | GET       |
| [/daemon/stop](#daemonstop-post)          | POST      |

#### /daemon/constants [GET]

returns the set of constants in use.

###### JSON Response
```javascript
{
  // Timestamp of the genesis block.
  "genesistimestamp": 1433600000, // Unix time
  // Maximum size, in bytes, of a block. Blocks larger than this will be
  // rejected by peers.
  "blocksizelimit": 2000000, // bytes
  // Target for how frequently new blocks should be mined.
  "blockfrequency": 600, // seconds per block
  // Height of the window used to adjust the difficulty.
  "targetwindow": 1000, // blocks
  // Duration of the window used to adjust the difficulty.
  "mediantimestampwindow": 11, // blocks
  // How far in the future a block can be without being rejected. A block
  // further into the future will not be accepted immediately, but the daemon
  // will attempt to accept the block as soon as it is valid.
  "futurethreshold": 10800, // seconds
  // Total number of blockstakes.
  "blockstakecount": "10000",
  // Number of children a block must have before it is considered "mature."
  "maturitydelay": 144, // blocks

  // Number of coins given to the miner of the first block. Note that elsewhere
  // in the API currency is typically returned in hastings and as a bignum.
  // This is not the case here.
  "initialcoinbase": 300000, // Coins.
  // Minimum number of coins paid out to the miner of a block (the coinbase
  // decreases with each block). Note that elsewhere in the API currency is
  // typically returned in hastings and as a bignum. This is not the case
  // here.
  "minimumcoinbase": 30000, // Coins

  // Initial target.
  "roottarget": [0,0,0,0,32,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  // Initial depth.
  "rootdepth": [255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255],

  // Largest allowed ratio between the old difficulty and the new difficulty.
  "maxadjustmentup": "5/2",
  // Smallest allowed ratio between the old difficulty and the new difficulty.
  "maxadjustmentdown": "2/5",

  // Number of Hastings in one coin.
  "onecoin": "1000000000000000000000000" // hastings per coin
}
```

#### /daemon/version [GET]

returns the version of the Rivine daemon currently running.

###### JSON Response
```javascript
{
  // Version number of the running Rivine Daemon. This number is visible to its
  // peers on the network.
  "version": "1.0.0"
}
```

#### /daemon/stop [STOP]

cleanly shuts down the daemon. May take a few seconds.

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).
