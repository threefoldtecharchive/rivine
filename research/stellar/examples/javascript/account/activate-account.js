const StellarSdk = require('stellar-sdk')
const server = new StellarSdk.Server('https://horizon-testnet.stellar.org')

async function fundAccount (sourceAccountKeypair, destinationAccountKeypair) {
  const fee = await server.fetchBaseFee();
  const account = await server.loadAccount(sourceAccountKeypair.publicKey());
  console.log(account)
  const createTransaction = new StellarSdk.TransactionBuilder(account, {
    fee,
    networkPassphrase: StellarSdk.Networks.TESTNET
  })
  .addOperation(StellarSdk.Operation.createAccount({
    destination: destinationAccountKeypair.publicKey(),
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

// Funds an address through friendbot using an account's public key
function fundThroughFriendbot (address) {
  return StellarSdk.HorizonAxiosClient.get("https://friendbot.stellar.org/?addr=" + address)
    .then(res => {
      if (res.status === 200) {
        console.log(`account with address: ${address} funded through friendbot!`)
      }
    })
    .catch(err => {
      console.log(err)
      return err
    })
}

async function fundAccountDemo (sourceKeypair, destinationKeypair, fundThroughFriendBot) {
  if (fundThroughFriendBot) {
    try {
      return await fundThroughFriendbot(destinationKeypair.publicKey())
    } catch (error) {
      return error
    }
  }

  return await fundAccount(sourceKeypair, destinationKeypair)
}

const sourceSecretKey = 'SBCSQ6QX2BRVF2HWYA3WVO2X7KP6S5P52AAEMEK2UO4WD6KC7RQI5ZST'
const sourceKeypair = StellarSdk.Keypair.fromSecret(sourceSecretKey);

const destinationSecretKey = 'SA7KHEGR2CE56IBIIRPVXJFGTXFOVL3KKJ2YZHBNNPME2EK5KIW3KT6O'
const destinationKeypair = StellarSdk.Keypair.fromSecret(destinationSecretKey)

fundAccountDemo(sourceKeypair, destinationKeypair, true)