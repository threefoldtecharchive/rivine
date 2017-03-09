Via this example people can input a file (ex. a pdf file) that will be hashed and added in a transaction to proof that a certain file is made before a timestamp.
The Application should be able to:

- receive a file , hash it, sign a transaction locally with this hash and send this transaction to some transaction builder nodes to put this proof on the chain. (fee will be needed for this transaction).
- Poll to see if the transaction is added to the chain and how deep (to know if the transaction is secured in the chain).
- Give a file and check if the file is included in the blockchain + when.
