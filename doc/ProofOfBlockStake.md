Proof Of BlockStake
===================

This document is meant to provide a good high level overview of the BlockStake
algorithm.

General
-------


protocol
--------

![POBSprotocoloverview](./POBSoverview.svg)


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
![POBSprotocoldifficulty](./POBSdifficulty.svg)
