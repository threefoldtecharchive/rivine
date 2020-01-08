const StellarSdk = require('stellar-sdk');

// The source account is the account we will be signing and sending from.
const sourceSecretKey = 'SBAHSEMGRJAOFGQRKAIC6TE4ZS2BQTIUM67Z6T7JRNSCQCKRMQJZBGDW';

const receiverPublicKey = 'GBXCKEIL3GTPY2LFUFYMDHUENSBM3HXKOIZM3N7PBMVA4OTEXU353AGM';

const server = new StellarSdk.Server('https://horizon-testnet.stellar.org');

async function transfer(fromSecret, toAddress) {
  const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceSecretKey);
  const sourcePublicKey = sourceKeypair.publicKey();
  const account = await server.loadAccount(sourcePublicKey);
  const fee = await server.fetchBaseFee();


  const transaction = new StellarSdk.TransactionBuilder(account, {
      fee,
      networkPassphrase: StellarSdk.Networks.TESTNET
    })
    // Add a payment operation to the transaction
    .addOperation(StellarSdk.Operation.payment({
      destination: receiverPublicKey,
      // The term native asset refers to lumens
      asset: StellarSdk.Asset.native(),
      amount: '15',
    }))
    .setTimeout(30)
    .build();

  transaction.sign(sourceKeypair);

  console.log(transaction.toEnvelope().toXDR('base64'));

  try {
    const transactionResult = await server.submitTransaction(transaction);
    console.log(JSON.stringify(transactionResult, null, 2));
    console.log('\nSuccess! View the transaction at: ');
    console.log(transactionResult._links.transaction.href);
  } catch (e) {
    console.log('An error has occured:');
    console.log(e);
  }
}

transfer(sourceSecretKey, receiverPublicKey)