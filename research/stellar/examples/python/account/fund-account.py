# pylint: disable=no-value-for-parameter
import requests
import sys, getopt
import click

@click.command()
@click.option('--address', help='Address to fund', required=True)

def fundThroughFriendbot(address):
  try:
    res = requests.get("https://friendbot.stellar.org/?addr=" + address)
    res.raise_for_status()
    print("account with address: {} funded through friendbot".format(address))
  except requests.exceptions.HTTPError:
    print(res.json())
    sys.exit(1)

if __name__ == '__main__':
  fundThroughFriendbot()