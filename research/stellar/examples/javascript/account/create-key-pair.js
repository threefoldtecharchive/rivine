const StellarSdk = require('stellar-sdk')

function createKeypair () {
  const keypair = StellarSdk.Keypair.random()
  console.log(`Key: ${keypair.secret()}`)
  console.log(`Address: ${keypair.publicKey()}`)
}

createKeypair()