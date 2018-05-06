package consensus

// TODO: enable and fix?!
/*
// testBlockSuite tests a wide variety of blocks.
func (cst *consensusSetTester) testBlockSuite() {
	//TODO: fix this
	// cst.testSimpleBlock()
	// cst.testSpendSiacoinsBlock()
	// cst.testSpendSiafunds()
}
*/
// // testSimpleBlock mines a simple block (no transactions except those
// // automatically added by the miner) and adds it to the consnesus set.
// func (cst *consensusSetTester) testSimpleBlock() {
// 	// Get the starting hash of the consenesus set.
// 	initialChecksum := cst.cs.dbConsensusChecksum()
// 	initialHeight := cst.cs.dbBlockHeight()
// 	initialBlockID := cst.cs.dbCurrentBlockID()
//
// 	// Mine and submit a block
// 	block, err := cst.miner.AddBlock()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Check that the consensus info functions changed as expected.
// 	resultingChecksum := cst.cs.dbConsensusChecksum()
// 	if initialChecksum == resultingChecksum {
// 		panic("checksum is unchanged after mining a block")
// 	}
// 	resultingHeight := cst.cs.dbBlockHeight()
// 	if resultingHeight != initialHeight+1 {
// 		panic("height of consensus set did not increase as expected")
// 	}
// 	currentPB := cst.cs.dbCurrentProcessedBlock()
// 	if currentPB.Block.ParentID != initialBlockID {
// 		panic("new processed block does not have correct information")
// 	}
// 	if currentPB.Block.ID() != block.ID() {
// 		panic("the state's current block is not reporting as the recently mined block.")
// 	}
// 	if currentPB.Height != initialHeight+1 {
// 		panic("the processed block is not reporting the correct height")
// 	}
// 	pathID, err := cst.cs.dbGetPath(currentPB.Height)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if pathID != block.ID() {
// 		panic("current path does not point to the correct block")
// 	}
//
// 	// Revert the block that was just added to the consensus set and check for
// 	// parity with the original state of consensus.
// 	parent, err := cst.cs.dbGetBlockMap(currentPB.Block.ParentID)
// 	if err != nil {
// 		panic(err)
// 	}
// 	_, _, err = cst.cs.dbForkBlockchain(parent)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if cst.cs.dbConsensusChecksum() != initialChecksum {
// 		panic("adding and reverting a block changed the consensus set")
// 	}
// 	// Re-add the block and check for parity with the first time it was added.
// 	// This test is useful because a different codepath is followed if the
// 	// diffs have already been generated.
// 	_, _, err = cst.cs.dbForkBlockchain(currentPB)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if cst.cs.dbConsensusChecksum() != resultingChecksum {
// 		panic("adding, reverting, and reading a block was inconsistent with just adding the block")
// 	}
// }
//
// // TestIntegrationSimpleBlock creates a consensus set tester and uses it to
// // call testSimpleBlock.
// func TestIntegrationSimpleBlock(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}
// 	t.Parallel()
// 	cst, err := createConsensusSetTester(t.Name())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cst.Close()
// 	cst.testSimpleBlock()
// }
//
// // testSpendSiacoinsBlock mines a block with a transaction spending siacoins
// // and adds it to the consensus set.
// func (cst *consensusSetTester) testSpendSiacoinsBlock() {
// 	// Create a random destination address for the output in the transaction.
// 	destAddr := randAddress()
//
// 	// Create a block containing a transaction with a valid siacoin output.
// 	txnValue := types.NewCurrency64(1200)
// 	txnBuilder := cst.wallet.StartTransaction()
// 	err := txnBuilder.FundSiacoins(txnValue)
// 	if err != nil {
// 		panic(err)
// 	}
// 	outputIndex := txnBuilder.AddSiacoinOutput(types.SiacoinOutput{Value: txnValue, UnlockHash: destAddr})
// 	txnSet, err := txnBuilder.Sign(true)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = cst.tpool.AcceptTransactionSet(txnSet)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Mine and apply the block to the consensus set.
// 	_, err = cst.miner.AddBlock()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// See that the destination output was created.
// 	outputID := txnSet[len(txnSet)-1].SiacoinOutputID(outputIndex)
// 	sco, err := cst.cs.dbGetSiacoinOutput(outputID)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if !sco.Value.Equals(txnValue) {
// 		panic("output added with wrong value")
// 	}
// 	if sco.UnlockHash != destAddr {
// 		panic("output sent to the wrong address")
// 	}
// }
//
// // TestIntegrationSpendSiacoinsBlock creates a consensus set tester and uses it
// // to call testSpendSiacoinsBlock.
// func TestIntegrationSpendSiacoinsBlock(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}
// 	t.Parallel()
// 	cst, err := createConsensusSetTester(t.Name())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cst.Close()
// 	cst.testSpendSiacoinsBlock()
// }
//
// // testSpendSiafunds spends siafunds on the blockchain.
// func (cst *consensusSetTester) testSpendSiafunds() {
// 	// Create a random destination address for the output in the transaction.
// 	destAddr := randAddress()
//
// 	// Create a block containing a transaction with a valid siafund output.
// 	txnValue := types.NewCurrency64(3)
// 	txnBuilder := cst.wallet.StartTransaction()
// 	err := txnBuilder.FundSiafunds(txnValue)
// 	if err != nil {
// 		panic(err)
// 	}
// 	outputIndex := txnBuilder.AddSiafundOutput(types.SiafundOutput{Value: txnValue, UnlockHash: destAddr})
// 	txnSet, err := txnBuilder.Sign(true)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = cst.tpool.AcceptTransactionSet(txnSet)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// Mine and apply the block to the consensus set.
// 	_, err = cst.miner.AddBlock()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	// See that the destination output was created.
// 	outputID := txnSet[len(txnSet)-1].SiafundOutputID(outputIndex)
// 	sfo, err := cst.cs.dbGetSiafundOutput(outputID)
// 	if err != nil {
// 		panic(err)
// 	}
// 	if sfo.Value.Cmp(txnValue) != 0 {
// 		panic("output added with wrong value")
// 	}
// 	if sfo.UnlockHash != destAddr {
// 		panic("output sent to the wrong address")
// 	}
//
// 	// Verify that all expected claims were created and added to the set of
// 	// delayed siacoin outputs.
// 	for i, id := range claimIDs {
// 		dsco, err := cst.cs.dbGetDSCO(cst.cs.dbBlockHeight()+types.MaturityDelay, id)
// 		if err != nil {
// 			panic(err)
// 		}
// 		if dsco.Value.Cmp(claimValues[i]) != 0 {
// 			panic("expected a different claim value on the siaclaim")
// 		}
// 	}
// }
//
// // TestIntegrationSpendSiafunds creates a consensus set tester and uses it
// // to call testSpendSiafunds.
// func (cst *consensusSetTester) TestIntegrationSpendSiafunds(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}
// 	t.Parallel()
// 	cst, err := createConsensusSetTester(t.Name())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cst.Close()
// 	cst.testSpendSiafunds()
// }
//
// // testDelayedOutputMaturity adds blocks that result in many delayed outputs
// // maturing at the same time, verifying that bulk maturity is handled
// // correctly.
//
// // TestRegressionDelayedOutputMaturity creates a consensus set tester and uses
// // it to call testDelayedOutputMaturity. In the past, bolt's ForEach function
// // had been used incorrectly resulting in the incorrect processing of bulk
// // delayed outputs.
//
// // testFileContractMaturity adds blocks that result in many file contracts
// // being closed at the same time.
//
// // TestRegressionFileContractMaturity creates a consensus set tester and uses
// // it to call testFileContractMaturity. In the past, bolt's ForEach function
// // had been used incorrectly, resulting in the incorrect processing of bulk
// // file contracts.
//
// /*
// // testPaymentChannelBlocks submits blocks to set up, use, and close a payment
// // channel.
// func (cst *consensusSetTester) testPaymentChannelBlocks() error {
// 	// The current method of doing payment channels is gimped because public
// 	// keys do not have timelocks. We will be hardforking to include timelocks
// 	// in public keys in 0.4.0, but in the meantime we need an alternate
// 	// method.
//
// 	// Gimped payment channels: 2-of-2 multisig where one key is controlled by
// 	// the funding entity, and one key is controlled by the receiving entity. An
// 	// address is created containing both keys, and then the funding entity
// 	// creates, but does not sign, a transaction sending coins to the channel
// 	// address. A second transaction is created that sends all the coins in the
// 	// funding output back to the funding entity. The receiving entity signs the
// 	// transaction with a timelocked signature. The funding entity will get the
// 	// refund after T blocks as long as the output is not double spent. The
// 	// funding entity then signs the first transaction and opens the channel.
// 	//
// 	// Creating the channel:
// 	//	1. Create a 2-of-2 unlock conditions, one key held by each entity.
// 	//	2. Funding entity creates, but does not sign, a transaction sending
// 	//		money to the payment channel address. (txn A)
// 	//	3. Funding entity creates and signs a transaction spending the output
// 	//		created in txn A that sends all the money back as a refund. (txn B)
// 	//	4. Receiving entity signs txn B with a timelocked signature, so that the
// 	//		funding entity cannot get the refund for several days. The funding entity
// 	//		is given a fully signed and eventually-spendable txn B.
// 	//	5. The funding entity signs and broadcasts txn A.
// 	//
// 	// Using the channel:
// 	//	Each the receiving entity and the funding entity keeps a record of how
// 	//	much has been sent down the unclosed channel, and watches the
// 	//	blockchain for a channel closing transaction. To send more money down
// 	//	the channel, the funding entity creates and signs a transaction sending
// 	//	X+y coins to the receiving entity from the channel address. The
// 	//	transaction is sent to the receiving entity, who will keep it and
// 	//	potentially sign and broadcast it later. The funding entity will only
// 	//	send money down the channel if 'work' or some other sort of event has
// 	//	completed that indicates the receiving entity should get more money.
// 	//
// 	// Closing the channel:
// 	//	The receiving entity will sign the transaction that pays them the most
// 	//	money and then broadcast that transaction. This will spend the output
// 	//	and close the channel, invalidating txn B and preventing any future
// 	//	transactions from being made over the channel. The channel must be
// 	//	closed before the timelock expires on the second signature in txn B,
// 	//	otherwise the funding entity will be able to get a full refund.
// 	//
// 	//	The funding entity should be waiting until either the receiving entity
// 	//	closes the channel or the timelock expires. If the receiving entity
// 	//	closes the channel, all is good. If not, then the funding entity can
// 	//	close the channel and get a full refund.
//
// 	// Create a 2-of-2 unlock conditions, 1 key for each the sender and the
// 	// receiver in the payment channel.
// 	sk1, pk1, err := crypto.StdKeyGen.Generate() // Funding entity.
// 	if err != nil {
// 		return err
// 	}
// 	sk2, pk2, err := crypto.StdKeyGen.Generate() // Receiving entity.
// 	if err != nil {
// 		return err
// 	}
// 	uc := types.UnlockConditions{
// 		PublicKeys: []types.SiaPublicKey{
// 			{
// 				Algorithm: types.SignatureEd25519,
// 				Key:       pk1[:],
// 			},
// 			{
// 				Algorithm: types.SignatureEd25519,
// 				Key:       pk2[:],
// 			},
// 		},
// 		SignaturesRequired: 2,
// 	}
// 	channelAddress := uc.UnlockHash()
//
// 	// Funding entity creates but does not sign a transaction that funds the
// 	// channel address. Because the wallet is not very flexible, the channel
// 	// txn needs to be fully custom. To get a custom txn, manually create an
// 	// address and then use the wallet to fund that address.
// 	channelSize := types.NewCurrency64(10e3)
// 	channelFundingSK, channelFundingPK, err := crypto.StdKeyGen.Generate()
// 	if err != nil {
// 		return err
// 	}
// 	channelFundingUC := types.UnlockConditions{
// 		PublicKeys: []types.SiaPublicKey{{
// 			Algorithm: types.SignatureEd25519,
// 			Key:       channelFundingPK[:],
// 		}},
// 		SignaturesRequired: 1,
// 	}
// 	channelFundingAddr := channelFundingUC.UnlockHash()
// 	fundTxnBuilder := cst.wallet.StartTransaction()
// 	if err != nil {
// 		return err
// 	}
// 	err = fundTxnBuilder.FundSiacoins(channelSize)
// 	if err != nil {
// 		return err
// 	}
// 	scoFundIndex := fundTxnBuilder.AddSiacoinOutput(types.SiacoinOutput{Value: channelSize, UnlockHash: channelFundingAddr})
// 	fundTxnSet, err := fundTxnBuilder.Sign(true)
// 	if err != nil {
// 		return err
// 	}
// 	fundOutputID := fundTxnSet[len(fundTxnSet)-1].SiacoinOutputID(int(scoFundIndex))
// 	channelTxn := types.Transaction{
// 		SiacoinInputs: []types.SiacoinInput{{
// 			ParentID:         fundOutputID,
// 			UnlockConditions: channelFundingUC,
// 		}},
// 		SiacoinOutputs: []types.SiacoinOutput{{
// 			Value:      channelSize,
// 			UnlockHash: channelAddress,
// 		}},
// 		TransactionSignatures: []types.TransactionSignature{{
// 			ParentID:       crypto.Hash(fundOutputID),
// 			PublicKeyIndex: 0,
// 			CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 		}},
// 	}
//
// 	// Funding entity creates and signs a transaction that spends the full
// 	// channel output.
// 	channelOutputID := channelTxn.SiacoinOutputID(0)
// 	refundUC, err := cst.wallet.NextAddress()
// 	refundAddr := refundUC.UnlockHash()
// 	if err != nil {
// 		return err
// 	}
// 	refundTxn := types.Transaction{
// 		SiacoinInputs: []types.SiacoinInput{{
// 			ParentID:         channelOutputID,
// 			UnlockConditions: uc,
// 		}},
// 		SiacoinOutputs: []types.SiacoinOutput{{
// 			Value:      channelSize,
// 			UnlockHash: refundAddr,
// 		}},
// 		TransactionSignatures: []types.TransactionSignature{{
// 			ParentID:       crypto.Hash(channelOutputID),
// 			PublicKeyIndex: 0,
// 			CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 		}},
// 	}
// 	sigHash := refundTxn.SigHash(0)
// 	cryptoSig1, err := crypto.SignHash(sigHash, sk1)
// 	if err != nil {
// 		return err
// 	}
// 	refundTxn.TransactionSignatures[0].Signature = cryptoSig1[:]
//
// 	// Receiving entity signs the transaction that spends the full channel
// 	// output, but with a timelock.
// 	refundTxn.TransactionSignatures = append(refundTxn.TransactionSignatures, types.TransactionSignature{
// 		ParentID:       crypto.Hash(channelOutputID),
// 		PublicKeyIndex: 1,
// 		Timelock:       cst.cs.dbBlockHeight() + 2,
// 		CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 	})
// 	sigHash = refundTxn.SigHash(1)
// 	cryptoSig2, err := crypto.SignHash(sigHash, sk2)
// 	if err != nil {
// 		return err
// 	}
// 	refundTxn.TransactionSignatures[1].Signature = cryptoSig2[:]
//
// 	// Funding entity will now sign and broadcast the funding transaction.
// 	sigHash = channelTxn.SigHash(0)
// 	cryptoSig0, err := crypto.SignHash(sigHash, channelFundingSK)
// 	if err != nil {
// 		return err
// 	}
// 	channelTxn.TransactionSignatures[0].Signature = cryptoSig0[:]
// 	err = cst.tpool.AcceptTransactionSet(append(fundTxnSet, channelTxn))
// 	if err != nil {
// 		return err
// 	}
// 	// Put the txn in a block.
// 	_, err = cst.miner.AddBlock()
// 	if err != nil {
// 		return err
// 	}
//
// 	// Try to submit the refund transaction before the timelock has expired.
// 	err = cst.tpool.AcceptTransactionSet([]types.Transaction{refundTxn})
// 	if err != types.ErrPrematureSignature {
// 		return err
// 	}
//
// 	// Create a transaction that has partially used the channel, and submit it
// 	// to the blockchain to close the channel.
// 	closeTxn := types.Transaction{
// 		SiacoinInputs: []types.SiacoinInput{{
// 			ParentID:         channelOutputID,
// 			UnlockConditions: uc,
// 		}},
// 		SiacoinOutputs: []types.SiacoinOutput{
// 			{
// 				Value:      channelSize.Sub(types.NewCurrency64(5)),
// 				UnlockHash: refundAddr,
// 			},
// 			{
// 				Value: types.NewCurrency64(5),
// 			},
// 		},
// 		TransactionSignatures: []types.TransactionSignature{
// 			{
// 				ParentID:       crypto.Hash(channelOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			},
// 			{
// 				ParentID:       crypto.Hash(channelOutputID),
// 				PublicKeyIndex: 1,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			},
// 		},
// 	}
// 	sigHash = closeTxn.SigHash(0)
// 	cryptoSig3, err := crypto.SignHash(sigHash, sk1)
// 	if err != nil {
// 		return err
// 	}
// 	closeTxn.TransactionSignatures[0].Signature = cryptoSig3[:]
// 	sigHash = closeTxn.SigHash(1)
// 	cryptoSig4, err := crypto.SignHash(sigHash, sk2)
// 	if err != nil {
// 		return err
// 	}
// 	closeTxn.TransactionSignatures[1].Signature = cryptoSig4[:]
// 	err = cst.tpool.AcceptTransactionSet([]types.Transaction{closeTxn})
// 	if err != nil {
// 		return err
// 	}
//
// 	// Mine the block with the transaction.
// 	_, err = cst.miner.AddBlock()
// 	if err != nil {
// 		return err
// 	}
// 	closeRefundID := closeTxn.SiacoinOutputID(0)
// 	closePaymentID := closeTxn.SiacoinOutputID(1)
// 	exists := cst.cs.db.inSiacoinOutputs(closeRefundID)
// 	if !exists {
// 		return errors.New("close txn refund output doesn't exist")
// 	}
// 	exists = cst.cs.db.inSiacoinOutputs(closePaymentID)
// 	if !exists {
// 		return errors.New("close txn payment output doesn't exist")
// 	}
//
// 	// Create a payment channel where the receiving entity never responds to
// 	// the initial transaction.
// 	{
// 		// Funding entity creates but does not sign a transaction that funds the
// 		// channel address. Because the wallet is not very flexible, the channel
// 		// txn needs to be fully custom. To get a custom txn, manually create an
// 		// address and then use the wallet to fund that address.
// 		channelSize := types.NewCurrency64(10e3)
// 		channelFundingSK, channelFundingPK, err := crypto.StdKeyGen.Generate()
// 		if err != nil {
// 			return err
// 		}
// 		channelFundingUC := types.UnlockConditions{
// 			PublicKeys: []types.SiaPublicKey{{
// 				Algorithm: types.SignatureEd25519,
// 				Key:       channelFundingPK[:],
// 			}},
// 			SignaturesRequired: 1,
// 		}
// 		channelFundingAddr := channelFundingUC.UnlockHash()
// 		fundTxnBuilder := cst.wallet.StartTransaction()
// 		err = fundTxnBuilder.FundSiacoins(channelSize)
// 		if err != nil {
// 			return err
// 		}
// 		scoFundIndex := fundTxnBuilder.AddSiacoinOutput(types.SiacoinOutput{Value: channelSize, UnlockHash: channelFundingAddr})
// 		fundTxnSet, err := fundTxnBuilder.Sign(true)
// 		if err != nil {
// 			return err
// 		}
// 		fundOutputID := fundTxnSet[len(fundTxnSet)-1].SiacoinOutputID(int(scoFundIndex))
// 		channelTxn := types.Transaction{
// 			SiacoinInputs: []types.SiacoinInput{{
// 				ParentID:         fundOutputID,
// 				UnlockConditions: channelFundingUC,
// 			}},
// 			SiacoinOutputs: []types.SiacoinOutput{{
// 				Value:      channelSize,
// 				UnlockHash: channelAddress,
// 			}},
// 			TransactionSignatures: []types.TransactionSignature{{
// 				ParentID:       crypto.Hash(fundOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			}},
// 		}
//
// 		// Funding entity creates and signs a transaction that spends the full
// 		// channel output.
// 		channelOutputID := channelTxn.SiacoinOutputID(0)
// 		refundUC, err := cst.wallet.NextAddress()
// 		refundAddr := refundUC.UnlockHash()
// 		if err != nil {
// 			return err
// 		}
// 		refundTxn := types.Transaction{
// 			SiacoinInputs: []types.SiacoinInput{{
// 				ParentID:         channelOutputID,
// 				UnlockConditions: uc,
// 			}},
// 			SiacoinOutputs: []types.SiacoinOutput{{
// 				Value:      channelSize,
// 				UnlockHash: refundAddr,
// 			}},
// 			TransactionSignatures: []types.TransactionSignature{{
// 				ParentID:       crypto.Hash(channelOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			}},
// 		}
// 		sigHash := refundTxn.SigHash(0)
// 		cryptoSig1, err := crypto.SignHash(sigHash, sk1)
// 		if err != nil {
// 			return err
// 		}
// 		refundTxn.TransactionSignatures[0].Signature = cryptoSig1[:]
//
// 		// Receiving entity never communitcates, funding entity must reclaim
// 		// the 'channelSize' coins that were intended to go to the channel.
// 		reclaimUC, err := cst.wallet.NextAddress()
// 		reclaimAddr := reclaimUC.UnlockHash()
// 		if err != nil {
// 			return err
// 		}
// 		reclaimTxn := types.Transaction{
// 			SiacoinInputs: []types.SiacoinInput{{
// 				ParentID:         fundOutputID,
// 				UnlockConditions: channelFundingUC,
// 			}},
// 			SiacoinOutputs: []types.SiacoinOutput{{
// 				Value:      channelSize,
// 				UnlockHash: reclaimAddr,
// 			}},
// 			TransactionSignatures: []types.TransactionSignature{{
// 				ParentID:       crypto.Hash(fundOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			}},
// 		}
// 		sigHash = reclaimTxn.SigHash(0)
// 		cryptoSig, err := crypto.SignHash(sigHash, channelFundingSK)
// 		if err != nil {
// 			return err
// 		}
// 		reclaimTxn.TransactionSignatures[0].Signature = cryptoSig[:]
// 		err = cst.tpool.AcceptTransactionSet(append(fundTxnSet, reclaimTxn))
// 		if err != nil {
// 			return err
// 		}
// 		block, _ := cst.miner.FindBlock()
// 		err = cst.cs.AcceptBlock(block)
// 		if err != nil {
// 			return err
// 		}
// 		reclaimOutputID := reclaimTxn.SiacoinOutputID(0)
// 		exists := cst.cs.db.inSiacoinOutputs(reclaimOutputID)
// 		if !exists {
// 			return errors.New("failed to reclaim an output that belongs to the funding entity")
// 		}
// 	}
//
// 	// Create a channel and the open the channel, but close the channel using
// 	// the timelocked signature.
// 	{
// 		// Funding entity creates but does not sign a transaction that funds the
// 		// channel address. Because the wallet is not very flexible, the channel
// 		// txn needs to be fully custom. To get a custom txn, manually create an
// 		// address and then use the wallet to fund that address.
// 		channelSize := types.NewCurrency64(10e3)
// 		channelFundingSK, channelFundingPK, err := crypto.StdKeyGen.Generate()
// 		if err != nil {
// 			return err
// 		}
// 		channelFundingUC := types.UnlockConditions{
// 			PublicKeys: []types.SiaPublicKey{{
// 				Algorithm: types.SignatureEd25519,
// 				Key:       channelFundingPK[:],
// 			}},
// 			SignaturesRequired: 1,
// 		}
// 		channelFundingAddr := channelFundingUC.UnlockHash()
// 		fundTxnBuilder := cst.wallet.StartTransaction()
// 		err = fundTxnBuilder.FundSiacoins(channelSize)
// 		if err != nil {
// 			return err
// 		}
// 		scoFundIndex := fundTxnBuilder.AddSiacoinOutput(types.SiacoinOutput{Value: channelSize, UnlockHash: channelFundingAddr})
// 		fundTxnSet, err := fundTxnBuilder.Sign(true)
// 		if err != nil {
// 			return err
// 		}
// 		fundOutputID := fundTxnSet[len(fundTxnSet)-1].SiacoinOutputID(int(scoFundIndex))
// 		channelTxn := types.Transaction{
// 			SiacoinInputs: []types.SiacoinInput{{
// 				ParentID:         fundOutputID,
// 				UnlockConditions: channelFundingUC,
// 			}},
// 			SiacoinOutputs: []types.SiacoinOutput{{
// 				Value:      channelSize,
// 				UnlockHash: channelAddress,
// 			}},
// 			TransactionSignatures: []types.TransactionSignature{{
// 				ParentID:       crypto.Hash(fundOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			}},
// 		}
//
// 		// Funding entity creates and signs a transaction that spends the full
// 		// channel output.
// 		channelOutputID := channelTxn.SiacoinOutputID(0)
// 		refundUC, err := cst.wallet.NextAddress()
// 		refundAddr := refundUC.UnlockHash()
// 		if err != nil {
// 			return err
// 		}
// 		refundTxn := types.Transaction{
// 			SiacoinInputs: []types.SiacoinInput{{
// 				ParentID:         channelOutputID,
// 				UnlockConditions: uc,
// 			}},
// 			SiacoinOutputs: []types.SiacoinOutput{{
// 				Value:      channelSize,
// 				UnlockHash: refundAddr,
// 			}},
// 			TransactionSignatures: []types.TransactionSignature{{
// 				ParentID:       crypto.Hash(channelOutputID),
// 				PublicKeyIndex: 0,
// 				CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 			}},
// 		}
// 		sigHash := refundTxn.SigHash(0)
// 		cryptoSig1, err := crypto.SignHash(sigHash, sk1)
// 		if err != nil {
// 			return err
// 		}
// 		refundTxn.TransactionSignatures[0].Signature = cryptoSig1[:]
//
// 		// Receiving entity signs the transaction that spends the full channel
// 		// output, but with a timelock.
// 		refundTxn.TransactionSignatures = append(refundTxn.TransactionSignatures, types.TransactionSignature{
// 			ParentID:       crypto.Hash(channelOutputID),
// 			PublicKeyIndex: 1,
// 			Timelock:       cst.cs.dbBlockHeight() + 2,
// 			CoveredFields:  types.CoveredFields{WholeTransaction: true},
// 		})
// 		sigHash = refundTxn.SigHash(1)
// 		cryptoSig2, err := crypto.SignHash(sigHash, sk2)
// 		if err != nil {
// 			return err
// 		}
// 		refundTxn.TransactionSignatures[1].Signature = cryptoSig2[:]
//
// 		// Funding entity will now sign and broadcast the funding transaction.
// 		sigHash = channelTxn.SigHash(0)
// 		cryptoSig0, err := crypto.SignHash(sigHash, channelFundingSK)
// 		if err != nil {
// 			return err
// 		}
// 		channelTxn.TransactionSignatures[0].Signature = cryptoSig0[:]
// 		err = cst.tpool.AcceptTransactionSet(append(fundTxnSet, channelTxn))
// 		if err != nil {
// 			return err
// 		}
// 		// Put the txn in a block.
// 		block, _ := cst.miner.FindBlock()
// 		err = cst.cs.AcceptBlock(block)
// 		if err != nil {
// 			return err
// 		}
//
// 		// Receiving entity never signs another transaction, so the funding
// 		// entity waits until the timelock is complete, and then submits the
// 		// refundTxn.
// 		for i := 0; i < 3; i++ {
// 			block, _ := cst.miner.FindBlock()
// 			err = cst.cs.AcceptBlock(block)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		err = cst.tpool.AcceptTransactionSet([]types.Transaction{refundTxn})
// 		if err != nil {
// 			return err
// 		}
// 		block, _ = cst.miner.FindBlock()
// 		err = cst.cs.AcceptBlock(block)
// 		if err != nil {
// 			return err
// 		}
// 		refundOutputID := refundTxn.SiacoinOutputID(0)
// 		exists := cst.cs.db.inSiacoinOutputs(refundOutputID)
// 		if !exists {
// 			return errors.New("timelocked refund transaction did not get spent correctly")
// 		}
// 	}
//
// 	return nil
// }
// */
//
// /*
// // TestPaymentChannelBlocks creates a consensus set tester and uses it to call
// // testPaymentChannelBlocks.
// func TestPaymentChannelBlocks(t *testing.T) {
// 	if testing.Short() {
// 		t.SkipNow()
// 	}
// 	cst, err := createConsensusSetTester(t.Name())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer cst.closeCst()
// 	err = cst.testPaymentChannelBlocks()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
// */
