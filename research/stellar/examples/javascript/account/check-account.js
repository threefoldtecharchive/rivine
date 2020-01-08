const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --address [string]')
    .demandOption(['address'])
    .argv;

async function checkAccountBalance (address) {
  account = await server.loadAccount(address)
  account.balances.forEach(balance => {
    console.log(balance)
  })
}

checkAccountBalance(argv.address)