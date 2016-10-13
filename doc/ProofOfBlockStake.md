Proof Of BlockStake
===================

This document is meant to provide a good high level overview of the BlockStake
algorithm.

General
-------


protocol
--------

![POBSprotocoloverview](https://rawgit.com/rivine/rivine/master/doc/POBSoverview.svg)


Stakemodifier
-------------


Transaction fees
----------------

For every transaction, a fee (exact amount to be determined) is charged.

Every Block starts with 3 fixed transactions:

All Fees from all transactions in the created block goes to:
25% to the BCN who generate the block
75% to address 0 (zero)
From address 0 , a certain percentage goes to BCN who generate the block
From address 1 , a fixed amount (ex 1BDG) goes to BCN who generate the block

Address 0 is filled by each transaction fee and distributed over the next few
block creators. Address 1 can be artificially filled by the foundation to start
up the system.

Since everything needs to be Backed, no new Digital Value can be created.
But there is still an incentive to mine because:

You get a certain amount of the fee of the transactions in the block.

If there are no transactions, you still get value from the fixed 2e transaction
which are fees from transactions in the previous blocks. And the BCN get also
value from the 3 fixed transaction from address 1.

If somebody makes a transaction error or does a test and send an amount to
address 0, that amount will not get lost but will be distributed over the BCNs.


Difficulty
----------

The hash in the POBS protocol results in a 256 bit integer so there are 2^256 possible combinations. If we want 1 block to be created every 10 minutes on average, this means that that 1 hash should match every 10\*60 seconds. The chance of having a match is also multiplied by the number of blockstakes you have available so for the starting difficulty this means it should be divided by the total number of blockstakes in the system. Difficulty is adjusted every 50 blocks to compensate for the fact that not every blockstake available is always participating in the POBS protocol.

![POBSprotocoldifficulty](https://rawgit.com/rivine/rivine/master/doc/POBSdifficulty.svg)
