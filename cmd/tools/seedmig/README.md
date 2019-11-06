# Seed Migration Tool

A small commandline tool to allow you to migrate from an old seed to the new seed format.
Prior to March 2018, Rivine used to use <https://github.com/NebulousLabs/entropy-mnemonics>,
the entropy-mnemonics system of Sia, the blockchain from which Rivine is forked.

These mnemonics use 29 words, instead of the 24 words that you get from (BIP-39](http://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki), for a 32 bytes entropy.
Words that originate from a different dictionary as well.

Very early users of Rivine-based blockchains, such as tfchain, might therefore still have such an old seed,
that they might wish to migrate to the new seed format, keeping the same entropy.

This is possible using this tool, from the tip of your fingers, or at least with a commandline under your control.

## Install

```
go install -u github.com/threefoldtech/rivine/cmd/tools/seedmig
```

## Usage Example

```
$ seedmig "across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen"
Input phrase (= seed): across knife thirsty puck itches hazard enmity fainted pebbles unzip echo queen rarest aphid bugs yanks okay abbey eskimos dove orange nouns august ailments inline rebel glass tyrant acumen
Input entropy: b37e7d01853d65bc9e0299a4ea821cf28e0a7c4dffd687b59a89c4dc9523612db4d1f919f777

✓ Input phrase and entropy checksum-verified

Output phrase (= seed): recall view document apology stone tattoo job farm pilot favorite mango topic thing dilemma dawn width marble proud pen meadow sing museum lucky present
```

In this example the user has the following "old" style mnmeonic:

```
across knife thirsty puck itches hazard enmity fainted pebbles
unzip echo queen rarest aphid bugs yanks okay abbey eskimos
dove orange nouns august ailments inline rebel glass tyrant
acumen
```

Which results in the entropy:

```
b37e7d01853d65bc9e0299a4ea821cf28e0a7c4dffd687b59a89c4dc9523612db4d1f919f777
```

This entropy however includes a checksum of 6 bytes added at the end, the last 12 hex-encoded
characters as can be seen above.

Therefore the actual entropy can be reduced to:

```
b37e7d01853d65bc9e0299a4ea821cf28e0a7c4dffd687b59a89c4dc9523612d
```

Which can be translated to words using the BIP-39 algorithm with as final result:

```
recall view document apology stone tattoo job farm pilot favorite mango
topic thing dilemma dawn width marble proud pen meadow sing museum lucky present
```

Your mnemonic in the "new" style, as used and supported by all existing wallet clients that
support Rivine chains, at the time of writing this document.
