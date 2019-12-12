# rivine-chain-explorer
A template Rivine Chain Explorer

## Development
- - -
### Installing Dependencies

```yarn install```
- - -
### Prerequisites

Export following environment variables in an `.env` file in the root of this project.

| Variable  | Default Value | Meaning | Required |
| ------------- | ------------- | ------------- | ------------- |
| VUE_APP_NAME  | None, Must be set!!  | Blockchain Name  | Yes |
| VUE_APP_API_URL  | https://explorer.testnet.nbh-digital.com  | API Url (daemon backend)  | No |
| VUE_APP_PRECISION  | 9  | Precision after decimal point for the currency | No |
| VUE_APP_UNIT  | GFT  | Unit that will be displayed  | No |

### Example env file content

```
VUE_APP_NAME=Goldchain
VUE_APP_PRECISION=9
VUE_APP_UNIT=GFT
```

### Configuration for local development

Run Caddy server in between frontend and backend.

```caddy -conf Caddyfile```

Export `VUE_APP_API_URL` as `http://localhost:2015`.


- - -
### Serving Frontend

```yarn run serve```