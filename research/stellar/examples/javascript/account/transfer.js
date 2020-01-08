const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --sourceKey [string] --destinationAddress [string] --amount [string]')
    .demandOption(['sourceKey', 'destinationAddress', 'amount'])
    .argv;


async function transfer(fromSecret, destination, amount) {
  const sourceKeypair = StellarSdk.Keypair.fromSecret(fromSecret)
  const account = await server.loadAccount(sourceKeypair.publicKey())

  const fee = await server.fetchBaseFee()

  const transaction = new StellarSdk.TransactionBuilder(account, {
      fee,
      networkPassphrase: StellarSdk.Networks.TESTNET
    })
    // Add a payment operation to the transaction
    .addOperation(StellarSdk.Operation.payment({
      destination,
      // The term native asset refers to lumens
      asset: StellarSdk.Asset.native(),
      amount: amount.toString(),
    }))
    .setTimeout(30)
    .build()

  transaction.sign(sourceKeypair)

  console.log(transaction.toEnvelope().toXDR('base64'))

  try {
    const transactionResult = await server.submitTransaction(transaction)
    console.log(JSON.stringify(transactionResult, null, 2))
    console.log('\nSuccess! View the transaction at: ')
    console.log(transactionResult._links.transaction.href)
  } catch (e) {
    console.log('An error has occured:')
    console.log(e)
  }
}

transfer(argv.sourceKey, argv.destinationAddress, argv.amount)