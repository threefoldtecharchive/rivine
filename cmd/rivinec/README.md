Rivinec Usage
==========

`rivinec` is the command line interface to Rivine, for use by power users and
those on headless servers. It comes as a part of the command line
package, and can be run as `./rivinec` from the same folder, or just by
calling `rivinec` if you move the binary into your path.

Most of the following commands have online help. For example, executing
`rivinec wallet send help` will list the arguments for that command,
while `rivinec host help` will list the commands that can be called
pertaining to hosting. `rivinec help` will list all of the top level
command groups that can be used.

You can change the address of where rivined is pointing using the `-a`
flag. For example, `rivinec -a :9000 status` will display the status of
the rivined instance launched on the local machine with `rivined -a :9000`.

Common tasks
------------
* `rivinec status` view block height

Wallet:
* `rivinec wallet init [-p]` initilize a wallet
* `rivinec wallet unlock` unlock a wallet
* `rivinec wallet status` retrieve wallet balance
* `rivinec wallet address` get a wallet address
* `rivinec wallet send [amount] [dest]` sends coin to an address

Full Descriptions
-----------------

#### Wallet tasks

* `rivinec wallet init [-p]` encrypts and initializes the wallet. If the
`-p` flag is provided, an encryption password is requested from the
user. Otherwise the initial seed is used as the encryption
password. The wallet must be initialized and unlocked before any
actions can be performed on the wallet.

Examples:
```bash
user@hostname:~$ rivinec -a :9920 wallet init
Seed is:
 cider sailor incur sober feast unhappy mundane sadness hinder aglow imitate amaze duties arrow gigantic uttered inflamed girth myriad jittery hexagon nail lush reef sushi pastry southern inkling acquire

Wallet encrypted with password: cider sailor incur sober feast unhappy mundane sadness hinder aglow imitate amaze duties arrow gigantic uttered inflamed girth myriad jittery hexagon nail lush reef sushi pastry southern inkling acquire
```

```bash
user@hostname:~$ rivinec -a :9920 wallet init -p
Wallet password:
Seed is:
 potato haunted fuming lordship library vane fever powder zippers fabrics dexterity hoisting emails pebbles each vampire rockets irony summon sailor lemon vipers foxes oneself glide cylinder vehicle mews acoustic

Wallet encrypted with given password
```

* `rivinec wallet unlock` prompts the user for the encryption password
to the wallet, supplied by the `init` command. The wallet must be
initialized and unlocked before any actions can take place.

* `rivinec wallet status` prints information about your wallet.

Example:
```bash
user@hostname:~$ rivinec wallet status
Wallet status:
Encrypted, Unlocked
Confirmed Balance:   61516458.00 ROC
Unconfirmed Balance: 64516461.00 ROC
Exact:               61516457999999999999999999999999 H
```

* `rivinec wallet address` returns a never seen before address for sending
coins to.

* `rivinec wallet send [amount] [dest]` Sends `amount` coins to
`dest`. `amount` is in the form X[.X] is a number expressed in a one coin unit,
which has a limited precision as indicated by the OneCoin config variable.

* `rivinec wallet lock` locks a wallet. After calling, the wallet must be unlocked
using the encryption password in order to use it further

* `rivinec wallet seeds` returns the list of secret seeds in use by the
wallet. These can be used to regenerate the wallet

* `rivinec wallet addseed` prompts the user for his encryption password,
as well as a new secret seed. The wallet will then incorporate this
seed into itself. This can be used for wallet recovery and merging.

#### Gateway tasks
* `rivinec gateway` prints info about the gateway, including its address and how
many peers it's connected to.

* `rivinec gateway list` prints a list of all currently connected peers.

* `rivinec gateway connect [address:port]` manually connects to a peer and adds it
to the gateway's node list.

* `rivinec gateway disconnect [address:port]` manually disconnects from a peer, but
leaves it in the gateway's node list.

#### Miner tasks
* `rivinec miner status` returns information about the miner. It is only
valid for when rivined is running.

* `rivinec miner start` starts running the CPU miner on one thread. This
is virtually useless outside of debugging.

* `rivinec miner stop` halts the CPU miner.

#### General commands
* `rivinec status` prints the current block ID, current block height, and
current target.

* `rivinec stop` sends the stop signal to rivined to safely terminate. This
has the same affect as C^c on the terminal.

* `rivinec version` displays the version string of rivinec.

* `rivinec update` checks the server for updates.
