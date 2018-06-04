# Beyond basic atomic swaps

## Problem
While atomic swaps are great and enable trustless trading, they do have some drawbacks:
1. slow ( It can take minutes for a trade to complete).
2. Transaction fees per chain can be quite high.
3. Current toolset requires wallets with full nodes.
4. Only penalties or built-in protection against  malicious users bailing out of the swap is the transaction fee cost.

Number 4 is not only annoying for other users but can also cost the other party transaction fees and lock their funds for a while. A naive orderbook implementation can easily be attacked this way.