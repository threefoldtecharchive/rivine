const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --sourceKey [string] --issuerAddress [string] --assetCode [string] --limit [string]')
    .demandOption(['sourceKey', 'issuerAddress', 'assetCode', 'limit'])
    .argv;

async function createTrusline(fromAccountSecret, issuerAddress, assetCode, limit) {
  const sourceKeypair = StellarSdk.Keypair.fromSecret(fromAccountSecret)
  const account = await server.loadAccount(sourceKeypair.publicKey())

  const fee = await server.fetchBaseFee()

  const asset = new StellarSdk.Asset(assetCode, issuerAddress)

  const transaction = new StellarSdk.TransactionBuilder(account, {
      fee,
      networkPassphrase: StellarSdk.Networks.TESTNET
    })
    // Add a payment operation to the transaction
    .addOperation(StellarSdk.Operation.changeTrust({
      asset,
      limit: limit.toString()
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

createTrusline(argv.sourceKey, argv.issuerAddress, argv.assetCode, argv.limit)