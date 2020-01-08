const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

const sourceSecretKey = 'SAU24VKHZHK24ABJF7AT6MMQDWNQRGCNCCIW7YDZVOMKYGI7HUWK6YI3'

const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceSecretKey);

console.log(sourceKeypair.secret())
console.log(sourceKeypair.publicKey())

// async function createAccount () {
//   const fee = await server.fetchBaseFee();

//   const createTransaction = new StellarSdk.TransactionBuilder(null, {
//     fee,
//     networkPassphrase: StellarSdk.Networks.TESTNET
//   })
//   .addOperation(StellarSdk.Operation.createAccount({
//     destination: '',
//     amount: '100'
//   }))
//   .build()
// }





function getAccountDetails (address) {
  server.loadAccount(address).then(res => {
    console.log(res)
  })
}