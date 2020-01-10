const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --sourceKey [string] --destinationAddress [string]')
    .demandOption(['sourceKey', 'destinationAddress'])
    .argv;

async function fundAccount (sourceKey, destination) {
  const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceKey)
  const account = await server.loadAccount(sourceKeypair.publicKey());

  const fee = await server.fetchBaseFee();

  const createTransaction = new StellarSdk.TransactionBuilder(account, {
    fee,
    networkPassphrase: StellarSdk.Networks.TESTNET
  })
  .addOperation(StellarSdk.Operation.createAccount({
    destination,
    startingBalance: '100'
  }))
  .setTimeout(30)
  .build()

  createTransaction.sign(sourceKeypair)

  console.log(createTransaction.toEnvelope().toXDR('base64'));

  try {
    const transactionResult = await server.submitTransaction(createTransaction);
    console.log(JSON.stringify(transactionResult, null, 2));
    console.log('\nSuccess! View the transaction at: ');
    console.log(transactionResult._links.transaction.href);
  } catch (e) {
    console.log('An error has occured:');
    console.log(e);
  }
}

fundAccount(argv.sourceKey, argv.destinationAddress)