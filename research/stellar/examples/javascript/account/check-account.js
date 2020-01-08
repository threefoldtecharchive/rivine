const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

async function checkAccountBalance (secret) {
  const keypair = StellarSdk.Keypair.fromSecret(secret);
  account = await server.loadAccount(keypair.publicKey())
  account.balances.forEach(balance => {
    console.log(balance)
  })
}

checkAccountBalance('SBCSQ6QX2BRVF2HWYA3WVO2X7KP6S5P52AAEMEK2UO4WD6KC7RQI5ZST')