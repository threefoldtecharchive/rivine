Rivined API
========

API calls return either JSON or no content. Success is indicated by 2xx HTTP
status codes, while errors are indicated by 4xx and 5xx HTTP status codes. If
an endpoint does not specify its expected status code refer to
[#standard-responses](#standard-responses).

There may be functional API calls which are not documented. These are not
guaranteed to be supported beyond the current release, and should not be used
in production.

Notes:
- Requests must set their User-Agent string to contain the substring "Rivine-Agent".
- By default, rivined listens on "localhost:23110". This can be changed using the
  `--api-addr` flag when running rivined.
- **Do not bind or expose the API to a non-loopback address unless you are
  aware of the possible dangers.**

Example GET curl call:
```
curl -A "Rivine-Agent" "localhost:23110/wallet/transactions?startheight=1&endheight=250"
```

Example POST curl call:
```
curl -A "Rivine-Agent" --data "amount=123&destination=abcd" "localhost:23110/wallet/coins"
```

Standard responses
------------------

#### Success

The standard response indicating the request was successfully processed is HTTP
status code `204 No Content`. If the request was successfully processed and the
server responded with JSON the HTTP status code is `200 OK`. Specific endpoints
may specify other 2xx status codes on success.

#### Error

The standard error response indicating the request failed for any reason, is a
4xx or 5xx HTTP status code with an error JSON object describing the error.
```javascript
{
    "message": String

    // There may be additional fields depending on the specific error.
}
```

Authentication
--------------

API authentication can be enabled with the `--authenticate-api` rivined flag.
Authentication is HTTP Basic Authentication as described in
[RFC 2617](https://tools.ietf.org/html/rfc2617), however, the username is the
empty string. The flag does not enforce authentication on all API endpoints.
Only endpoints that expose sensitive information or modify state require
authentication.

For example, if the API password is "foobar" the request header should include
```
Authorization: Basic OmZvb2Jhcg==
```

Units
-----

Unless otherwise specified, all parameters should be specified in their
smallest possible unit. For example, size should always be specified in bytes
and Coins should be specified in hastings. JSON values returned by the API
will also use the smallest possible unit, unless otherwise specified.

If a numbers is returned as a string in JSON, it should be treated as an
arbitrary-precision number (bignum), and it should be parsed with your
language's corresponding bignum library. Currency values are the most common
example where this is necessary.

Table of contents
-----------------

- [Daemon](#daemon)
- [Consensus](#consensus)
- [Gateway](#gateway)- [Wallet](#wallet)

Daemon
------

| Route                                     | HTTP verb |
| ----------------------------------------- | --------- |
| [/daemon/constants](#daemonconstants-get) | GET       |
| [/daemon/version](#daemonversion-get)     | GET       |
| [/daemon/stop](#daemonstop-post)          | POST      |

For examples and detailed descriptions of request and response parameters,
refer to [Daemon.md](/doc/api/Daemon.md).

#### /daemon/constants [GET]

returns the set of constants in use.

###### JSON Response [(with comments)](/doc/api/Daemon.md#json-response)
```javascript
{
  "genesistimestamp":      1257894000, // Unix time
  "blocksizelimit":        2000000,    // bytes
  "blockfrequency":        600,        // seconds per block
  "targetwindow":          1000,       // blocks
  "mediantimestampwindow": 11,         // blocks
  "futurethreshold":       10800,      // seconds
  "siafundcount":          "10000",
  "siafundportion":        "39/1000",
  "maturitydelay":         144,        // blocks

  "initialcoinbase": 300000, // SiaCoins (see note in Daemon.md)
  "minimumcoinbase": 30000,  // SiaCoins (see note in Daemon.md)

  "roottarget": [0,0,0,0,32,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],
  "rootdepth":  [255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255,255],

  "maxadjustmentup":   "5/2",
  "maxadjustmentdown": "2/5",

  "onecoin": "1000000000000000000000000" // hastings per siacoin
}
```

#### /daemon/version [GET]

returns the version of the Sia daemon currently running.

###### JSON Response [(with comments)](/doc/api/Daemon.md#json-response-1)
```javascript
{
  "version": "1.0.0"
}
```

#### /daemon/stop [POST]

cleanly shuts down the daemon. May take a few seconds.

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

Consensus
---------

| Route                        | HTTP verb |
| ---------------------------- | --------- |
| [/consensus](#consensus-get) | GET       |

For examples and detailed descriptions of request and response parameters,
refer to [Consensus.md](/doc/api/Consensus.md).

#### /consensus [GET]

returns information about the consensus set, such as the current block height.

###### JSON Response [(with comments)](/doc/api/Consensus.md#json-response)
```javascript
{
  "synced":       true,
  "height":       62248,
  "currentblock": "00000000000008a84884ba827bdc868a17ba9c14011de33ff763bd95779a9cf1",
  "target":       [0,0,0,0,0,0,11,48,125,79,116,89,136,74,42,27,5,14,10,31,23,53,226,238,202,219,5,204,38,32,59,165]
}
```

Gateway
-------

| Route                                                                              | HTTP verb |
| ---------------------------------------------------------------------------------- | --------- |
| [/gateway](#gateway-get-example)                                                   | GET       |
| [/gateway/connect/___:netaddress___](#gatewayconnectnetaddress-post-example)       | POST      |
| [/gateway/disconnect/___:netaddress___](#gatewaydisconnectnetaddress-post-example) | POST      |

For examples and detailed descriptions of request and response parameters,
refer to [Gateway.md](/doc/api/Gateway.md).

#### /gateway [GET] [(example)](/doc/api/Gateway.md#gateway-info)

returns information about the gateway, including the list of connected peers.

###### JSON Response [(with comments)](/doc/api/Gateway.md#json-response)
```javascript
{
    "netaddress": String,
    "peers":      []{
        "netaddress": String,
        "version":    String,
        "inbound":    Boolean
    }
}
```

#### /gateway/connect/___:netaddress___ [POST] [(example)](/doc/api/Gateway.md#connecting-to-a-peer)

connects the gateway to a peer. The peer is added to the node list if it is not
already present. The node list is the list of all nodes the gateway knows
about, but is not necessarily connected to.

###### Path Parameters [(with comments)](/doc/api/Gateway.md#path-parameters)
```
:netaddress
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

#### /gateway/disconnect/___:netaddress___ [POST] [(example)](/doc/api/Gateway.md#disconnecting-from-a-peer)

disconnects the gateway from a peer. The peer remains in the node list.

###### Path Parameters [(with comments)](/doc/api/Gateway.md#path-parameters-1)
```
:netaddress
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

TransactionPool
---------------

| Route                                                           | HTTP verb |
| --------------------------------------------------------------- | --------- |
| [/transactionpool/transactions](#transactions-post)             | POST      |


#### /transactionpool/transactions [POST]

Provide an externally constructed and signed transaction to the transactionpool.

###### JSON BODY

```javascript
{
  "coininputs": [
    {
      "parentid": "13b157d7e1bb8452c385acc39aa2e0f4d3dc982aa6ca2802dc43a2535b02bfb9",
      "unlockconditions": {
        "timelock": 0,
        "publickeys": [
          {
            "algorithm": "ed25519",
            "key": "gunn3wmyVZZqza4PwTdhPlZ0ttiEONuu1V+Q0OAdccU="
          }
        ],
        "signaturesrequired": 1
      }
    }
  ],
  "coinoutputs": [
    {
      "value": "120000000000000000000000000",
      "unlockhash": "354a92fda2ee24cd8bb6d588aa4c670325a1c226cab4b13a4b62fac154656ee5398532e42c9e"
    }
  ],
  "blockstakeinputs": null,
  "blockstakeoutputs": null,
  "minerfees": [
    "10000000000000000000000000"
  ],
  "arbitrarydata": null,
  "transactionsignatures": [
    {
      "parentid": "13b157d7e1bb8452c385acc39aa2e0f4d3dc982aa6ca2802dc43a2535b02bfb9",
      "publickeyindex": 0,
      "timelock": 0,
      "coveredfields": {
        "wholetransaction": true,
        "coininputs": null,
        "coinoutputs": null,
        "blockstakeinputs": null,
        "blockstakeoutputs": null,
        "minerfees": null,
        "arbitrarydata": null,
        "transactionsignatures": null
      },
      "signature": "1/zzGCdDzII2kv2y3+9Roq5p9sxokgiikXT4HdEkw3cbq9SMnXLRiYJfp2FcXSg1Hqk3OsJcAREgBxgg9fBQBg=="
    }
  ]
}
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).


Wallet
------

| Route                                                           | HTTP verb |
| --------------------------------------------------------------- | --------- |
| [/wallet](#wallet-get)                                          | GET       |
| [/wallet/033x](#wallet033x-post)                                | POST      |
| [/wallet/address](#walletaddress-get)                           | GET       |
| [/wallet/addresses](#walletaddresses-get)                       | GET       |
| [/wallet/backup](#walletbackup-get)                             | GET       |
| [/wallet/init](#walletinit-post)                                | POST      |
| [/wallet/lock](#walletlock-post)                                | POST      |
| [/wallet/seed](#walletseed-post)                                | POST      |
| [/wallet/seeds](#walletseeds-get)                               | GET       |
| [/wallet/coins](#walletcoins-post)                              | POST      |
| [/wallet/blockstakes](#walletblockstakes-post)                  | POST      |
| [/wallet/data](#walletdata-post)                                | POST      |
| [/wallet/siagkey](#walletsiagkey-post)                          | POST      |
| [/wallet/transaction/___:id___](#wallettransactionid-get)       | GET       |
| [/wallet/transactions](#wallettransactions-get)                 | GET       |
| [/wallet/transactions/___:addr___](#wallettransactionsaddr-get) | GET       |
| [/wallet/unlock](#walletunlock-post)                            | POST      |

For examples and detailed descriptions of request and response parameters,
refer to [Wallet.md](/doc/api/Wallet.md).

#### /wallet [GET]

returns basic information about the wallet, such as whether the wallet is
locked or unlocked.

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response)
```javascript
{
  "encrypted": true,
  "unlocked":  true,

  "confirmedsiacoinbalance":     "123456", // hastings, big int
  "unconfirmedoutgoingsiacoins": "0",      // hastings, big int
  "unconfirmedincomingsiacoins": "789",    // hastings, big int

  "siafundbalance":      "1",    // siafunds, big int
  "siacoinclaimbalance": "9001", // hastings, big int
}
```

#### /wallet/address [GET]

gets a new address from the wallet generated by the primary seed. An error will
be returned if the wallet is locked.

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-1)
```javascript
{
  "address": "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab"
}
```

#### /wallet/addresses [GET]

fetches the list of addresses from the wallet.

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-2)
```javascript
{
  "addresses": [
    "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab",
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  ]
}
```

#### /wallet/backup [GET]

creates a backup of the wallet settings file. Though this can easily be done
manually, the settings file is often in an unknown or difficult to find
location. The /wallet/backup call can spare users the trouble of needing to
find their wallet file.

###### Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-1)
```
destination
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

#### /wallet/init [POST]

initializes the wallet. After the wallet has been initialized once, it does not
need to be initialized again, and future calls to /wallet/init will return an
error. The encryption password is provided by the api call. If the password is
blank, then the password will be set to the same as the seed.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-2)
```
encryptionpassword
dictionary // Optional, default is english.
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-3)
```javascript
{
  "primaryseed": "hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello"
}
```

#### /wallet/seed [POST]

gives the wallet a seed to track when looking for incoming transactions. The
wallet will be able to spend outputs related to addresses created by the seed.
The seed is added as an auxiliary seed, and does not replace the primary seed.
Only the primary seed will be used for generating new addresses.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-3)
```
encryptionpassword
dictionary
seed
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

#### /wallet/seeds [GET]

returns the list of seeds in use by the wallet. The primary seed is the only
seed that gets used to generate new addresses. This call is unavailable when
the wallet is locked.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-4)
```
dictionary
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-4)
```javascript
{
  "primaryseed":        "hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello",
  "addressesremaining": 2500,
  "allseeds":           [
    "hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello world hello",
    "foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo",
  ]
}
```

#### /wallet/coins [POST]

sends siacoins to an address. The outputs are arbitrarily selected from
addresses in the wallet.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-5)
```
amount      // hastings
destination // address
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-5)
```javascript
{
  "transactionids": [
    "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  ]
}
```

#### /wallet/blockstakes [POST]

sends siafunds to an address. The outputs are arbitrarily selected from
addresses in the wallet. Any siacoins available in the siafunds being sent (as
well as the siacoins available in any siafunds that end up in a refund address)
will become available to the wallet as siacoins after 144 confirmations. To
access all of the siacoins in the siacoin claim balance, send all of the
siafunds to an address in your control (this will give you all the siacoins,
while still letting you control the siafunds).

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-6)
```
amount      // siafunds
destination // address
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-6)
```javascript
{
  "transactionids": [
    "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  ]
}
```

#### /wallet/data [POST]

Registers data on the blockchain. A transaction is created which sends the
minimal amount of 1 hasting to the provided address. The data provided is added
as arbitrary data in the transaction

###### Query String Parameters
```
// Address that is receiving the 1 hasting sent in the transaction
destination     // address

// The base64 encoded representation of the data
data            // base64 string
```

###### JSON Response
```javascript
{
  // Array of IDs of the transactions that were created when sending the coins.
  // The last transaction contains the output headed to the 'destination'.
  // Transaction IDs are 64 character long hex strings.
  "transactionids": [
    "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
  ]
}
```

#### /wallet/siagkey [POST]

loads a key into the wallet that was generated by siag. Most siafunds are
currently in addresses created by siag.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-7)
```
encryptionpassword
keyfiles
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

#### /wallet/lock [POST]

locks the wallet, wiping all secret keys. After being locked, the keys are
encrypted. Queries for the seed, to send siafunds, and related queries become
unavailable. Queries concerning transaction history and balance are still
available.

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).

#### /wallet/transaction/___:id___ [GET]

gets the transaction associated with a specific transaction id.

###### Path Parameters [(with comments)](/doc/api/Wallet.md#path-parameters)
```
:id
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-7)
```javascript
{
  "transaction": {
    "transaction": {
      // See types.Transaction in https://github.com/rivine/rivine/blob/master/types/transactions.go
    },
    "transactionid":         "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
    "confirmationheight":    50000,
    "confirmationtimestamp": 1257894000,
    "inputs": [
      {
        "fundtype":       "siacoin input",
        "walletaddress":  false,
        "relatedaddress": "1234567890abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ab",
        "value":          "1234", // hastings or siafunds, depending on fundtype, big int
      }
    ],
    "outputs": [
      {
        "fundtype":       "siacoin output",
        "maturityheight": 50000,
        "walletaddress":  false,
        "relatedaddress": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
        "value":          "1234", // hastings or siafunds, depending on fundtype, big int
      }
    ]
  }
}
```

#### /wallet/transactions [GET]

returns a list of transactions related to the wallet in chronological order.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-8)
```
startheight // block height
endheight   // block height
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-8)
```javascript
{
  "confirmedtransactions": [
    {
      // See the documentation for '/wallet/transaction/:id' for more information.
    }
  ],
  "unconfirmedtransactions": [
    {
      // See the documentation for '/wallet/transaction/:id' for more information.
    }
  ]
}
```

#### /wallet/transactions/___:addr___ [GET]

returns all of the transactions related to a specific address.

###### Path Parameters [(with comments)](/doc/api/Wallet.md#path-parameters-1)
```
:addr
```

###### JSON Response [(with comments)](/doc/api/Wallet.md#json-response-9)
```javascript
{
  "transactions": [
    {
      // See the documentation for '/wallet/transaction/:id' for more information.
    }
  ]
}
```

#### /wallet/unlock [POST]

unlocks the wallet. The wallet is capable of knowing whether the correct
password was provided.

###### Query String Parameters [(with comments)](/doc/api/Wallet.md#query-string-parameters-9)
```
encryptionpassword
```

###### Response
standard success or error response. See
[#standard-responses](#standard-responses).
