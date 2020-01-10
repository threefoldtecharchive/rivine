# pylint: disable=no-value-for-parameter
from stellar_sdk import Server, Keypair, TransactionBuilder, Network
from stellar_sdk.exceptions import BadRequestError
import click

@click.command()
@click.option('--sourcekey', help='Secret key of the source', required=True)
@click.option('--issueraddress', help='Destination address of the account to activate', required=True)
@click.option('--assetcode', required=True)
@click.option('--limit')

def create_trustline(sourcekey, issueraddress, assetcode, limit):
  server = Server(horizon_url="https://horizon-testnet.stellar.org")
  source_keypair = Keypair.from_secret(sourcekey)
  source_public_key = source_keypair.public_key
  source_account = server.load_account(source_public_key)

  base_fee = server.fetch_base_fee()

  transaction = (
      TransactionBuilder(
          source_account=source_account,
          network_passphrase=Network.TESTNET_NETWORK_PASSPHRASE,
          base_fee=base_fee,
      )
          .append_change_trust_op(asset_issuer=issueraddress, limit=limit, asset_code=assetcode)
          .set_timeout(30)
          .build()
  )

  transaction.sign(source_keypair)

  print(transaction.to_xdr())

  try:
    response = server.submit_transaction(transaction)
    print("Transaction hash: {}".format(response["hash"]))
    print(response)
  except BadRequestError as e:
    print(e)

if __name__ == '__main__':
  create_trustline()