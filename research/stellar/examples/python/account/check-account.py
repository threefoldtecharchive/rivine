# pylint: disable=no-value-for-parameter
from stellar_sdk import Server
import click

@click.command()
@click.option('--address', help='Address to check', required=True)

def check_account(address):
  
  server = Server(horizon_url="https://horizon-testnet.stellar.org")
  response= server.accounts().account_id(address).call()
  print(response['balances'])

if __name__ == '__main__':
  check_account()