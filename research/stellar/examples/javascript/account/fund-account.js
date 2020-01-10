const StellarSdk = require('stellar-sdk')

var argv = require('yargs')
    .usage('Usage: $0 --address [string]')
    .demandOption(['address'])
    .argv;


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


fundThroughFriendbot(argv.address)