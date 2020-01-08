const axios = require('axios')
const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

const sourceSecretKey = 'SBCSQ6QX2BRVF2HWYA3WVO2X7KP6S5P52AAEMEK2UO4WD6KC7RQI5ZST'
const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceSecretKey);

const destinationSecretKey = 'SAU24VKHZHK24ABJF7AT6MMQDWNQRGCNCCIW7YDZVOMKYGI7HUWK6YI3'
const destinationKeypair = StellarSdk.Keypair.fromSecret(destinationSecretKey)

async function createAccount () {
  const fee = await server.fetchBaseFee();
  const account = await server.loadAccount(sourceKeypair.publicKey());
  console.log(account)
  const createTransaction = new StellarSdk.TransactionBuilder(account, {
    fee,
    networkPassphrase: StellarSdk.Networks.TESTNET
  })
  .addOperation(StellarSdk.Operation.createAccount({
    destination: destinationKeypair.publicKey(),
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

// Funds an address through friendbot
function fundThroughFriendbot (address) {
  axiosClient = StellarSdk.HorizonAxiosClient
  axiosClient.get("https://friendbot.stellar.org/?addr=" + address)
    .then(res => {
      console.log(res)
    })
    .catch(err => {
      console.log(err)
    })
}

createAccount()