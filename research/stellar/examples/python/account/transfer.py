# pylint: disable=no-value-for-parameter
from stellar_sdk import Server, Keypair, TransactionBuilder, Network
from stellar_sdk.exceptions import BadRequestError
import click

@click.command()
@click.option('--sourcekey', help='Secret key of the source', required=True)
@click.option('--destinationaddress', help='Destination address of the account to activate', required=True)
@click.option('--amount', help='Amount to transfer', required=True, type=int)
@click.option('--asset', help='Asset format code:issuer', default='')

def transfer(sourcekey, destinationaddress, amount, asset):
  server = Server(horizon_url="https://horizon-testnet.stellar.org")
  source_keypair = Keypair.from_secret(sourcekey)
  source_public_key = source_keypair.public_key
  source_account = server.load_account(source_public_key)

  base_fee = server.fetch_base_fee()

  issuer = None

  if asset == '':
    asset = 'XLM'
  else:
    assetStr = asset.split(':')
    if len(assetStr) != 2:
      return Exception('Wrong asset format')
    asset = assetStr[0]
    issuer = assetStr[1]

  transaction = (
      TransactionBuilder(
          source_account=source_account,
          network_passphrase=Network.TESTNET_NETWORK_PASSPHRASE,
          base_fee=base_fee,
      )
          .append_payment_op(destination=destinationaddress, amount=str(amount), asset_code=asset, asset_issuer=issuer)
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
  transfer()