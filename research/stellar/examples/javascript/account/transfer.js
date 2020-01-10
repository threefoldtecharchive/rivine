const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --sourceKey [string] --destinationAddress [string] --amount [string] --asset format: `code:issuer`')
    .demandOption(['sourceKey', 'destinationAddress', 'amount'])
    .option('asset', {
      alias: 'a',
      type: 'string',
      description: 'format: code:issuer',
      default: ''
    })
    .argv;


async function transfer(sourceKey, destination, amount, asset) {
  const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceKey)
  const account = await server.loadAccount(sourceKeypair.publicKey())

  const fee = await server.fetchBaseFee()

  if (asset === '') {
    asset = StellarSdk.Asset.native()
  } else {
    assetStr = asset.split(':')
    if (assetStr.length !== 2) throw new Error('asset has wrong format')
    asset = new StellarSdk.Asset(assetStr[0], assetStr[1])
  }

  const transaction = new StellarSdk.TransactionBuilder(account, {
      fee,
      networkPassphrase: StellarSdk.Networks.TESTNET
    })
    // Add a payment operation to the transaction
    .addOperation(StellarSdk.Operation.payment({
      destination,
      // The term native asset refers to lumens
      asset,
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

transfer(argv.sourceKey, argv.destinationAddress, argv.amount, argv.asset)