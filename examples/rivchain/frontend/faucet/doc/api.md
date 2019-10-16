# API documentation

Following is a brief description about the api endpoints available on the faucet,
the expected bodies, and the expected responses.

## Request coins

endpoint: `/api/v1/coins`
method: `POST`

### Request body

type: `application/json`
data: 

```json
{
	"address": "UnlockHash string",
	"amount": "Amount of tokens (unsigned int)| optional (default 300)"
}
```

### Response body

type: `application/json`
data:

```json
{
	"txid": "Transaction ID"
}
```

## Authorize address

endpoint: `/api/v1/authorize`
method: `POST`

### Request body

type: `application/json`
data:

```json
{
	"address": "UnlockHash string"
}
```

### Response body

type: `application/json`
data:

```json
{
	"txid": "Transaction ID"
}
```

## Deauthorize address

endpoint: `/api/v1/deauthorize`
method: `POST`

### Request body

type: `application/json`
data:

```json
{
	"address": "UnlockHash string"
}
```

### Response body

type: `application/json`
data:

```json
{
	"txid": "Transaction ID"
}
```
