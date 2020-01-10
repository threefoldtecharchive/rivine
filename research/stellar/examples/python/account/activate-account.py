# pylint: disable=no-value-for-parameter
from stellar_sdk import TransactionBuilder, Server, Network, Keypair
from stellar_sdk.exceptions import BadRequestError
import click

@click.command()
@click.option('--sourcekey', help='Secret key of the source', required=True)
@click.option('--destinationaddress', help='Destination address of the account to activate', required=True)

def activate_account(sourcekey, destinationaddress):
  server = Server(horizon_url="https://horizon-testnet.stellar.org")
  source = Keypair.from_secret(sourcekey)

  source_account = server.load_account(account_id=source.public_key)
  transaction = TransactionBuilder(
      source_account=source_account,
      network_passphrase=Network.TESTNET_NETWORK_PASSPHRASE,
      base_fee=100) \
      .append_create_account_op(destination=destinationaddress, starting_balance="12.25") \
      .build()
  transaction.sign(source)
  try:
    response = server.submit_transaction(transaction)
    print("Transaction hash: {}".format(response["hash"]))
    print(response)
  except BadRequestError as e:
    print(e)

if __name__ == '__main__':
  activate_account()