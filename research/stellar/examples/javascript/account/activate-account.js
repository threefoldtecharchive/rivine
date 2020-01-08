const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

var argv = require('yargs')
    .usage('Usage: $0 --sourceKey [string] --destinationAddress [string]')
    .demandOption(['sourceKey', 'destinationAddress'])
    .argv;

async function fundAccount (sourceAccountKeypair, destination) {
  const fee = await server.fetchBaseFee();
  const account = await server.loadAccount(sourceAccountKeypair.publicKey());
  console.log(account)
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

  createTransaction.sign(sourceAccountKeypair)

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

fundAccount(StellarSdk.Keypair.fromSecret(argv.sourceKey), argv.destinationAddress)