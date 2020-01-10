# pylint: disable=no-value-for-parameter
from stellar_base import Address
import click

@click.command()
@click.option('--address', help='Address to check', required=True)

def check_account(address):
  address = Address(address=address)
  address.get()

  print('Balances: {}'.format(address.balances))
  print('Sequence Number: {}'.format(address.sequence))
  print('Flags: {}'.format(address.flags))
  print('Signers: {}'.format(address.signers))
  print('Data: {}'.format(address.data))

if __name__ == '__main__':
  check_account()