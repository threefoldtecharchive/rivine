package consensus

// TODO: ENABLE AND FIX:
/*
// TODO: CalculateStakeModifier seems to have been moved to the consensusset
// Moved in: https://github.com/threefoldtech/rivine/commit/473885476ccbcc6c259fa55e3af2aa818fe26db6#diff-4a95df743bd21a9b70774bb3981eac9a
func TestCalculateStakeModifier(t *testing.T) {

	alloneshash := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
	cs := consensusSetStub{}
	types.StakeModifierDelay = 2000

	persist := persistence{Height: types.BlockHeight(2254)} //this blockheight is the current blockheight
	bc := BlockCreator{cs: cs, persist: persist}
	stakemodifier := bc.CalculateStakeModifier()
	if stakemodifier.Cmp(alloneshash) != 0 {
		t.Error("The stakemodifier should be all binary 1's in this specific case")
	}

	persist2 := persistence{Height: types.BlockHeight(2510)} //
	bc2 := BlockCreator{cs: cs, persist: persist2}
	stakemodifier2 := bc2.CalculateStakeModifier()
	if stakemodifier2.Cmp(alloneshash) != 0 {
		t.Error("The stakemodifier should be all binary 1's in this specific case")
	}

	stakemodifierblockheight0 := big.NewInt(0)
	stakemodifierblockheight0.SetString("3C6215E851515D3CBABBBD19002F7E6ABCFCE86CDA4856414165071590AB394", 16)
	persist3 := persistence{Height: types.BlockHeight(0)}
	bc3 := BlockCreator{cs: cs, persist: persist3}
	stakemodifier3 := bc3.CalculateStakeModifier()
	if stakemodifier3.Cmp(stakemodifierblockheight0) != 0 {
		t.Error("The stakemodifier is deterministic for blockheight 0 and StakeModifierDelay 2000 because the blockmodifier is calculated with deterministic blockIDs pre genesis")
	}

	stakemodifierblockheight345 := big.NewInt(0)
	stakemodifierblockheight345.SetString("E7D660EB51B8872A26173D6BC280F8C308E626B1C6CF3F3DBCEF4D9BA6D78AFD", 16)
	persist4 := persistence{Height: types.BlockHeight(345)}
	bc4 := BlockCreator{cs: cs, persist: persist4}
	stakemodifier4 := bc4.CalculateStakeModifier()
	if stakemodifier4.Cmp(stakemodifierblockheight345) != 0 {
		t.Error("The stakemodifier is deterministic for blockheight 345 and StakeModifierDelay 2000 because the blockmodifier is calculated with deterministic blockIDs pre genesis")
	}

	persist5 := persistence{Height: types.BlockHeight(12345)}
	bc5 := BlockCreator{cs: cs, persist: persist5}
	stakemodifier5 := bc5.CalculateStakeModifier()
	persist6 := persistence{Height: types.BlockHeight(12345 + 256)}
	bc6 := BlockCreator{cs: cs, persist: persist6}
	stakemodifier6 := bc6.CalculateStakeModifier()
	if stakemodifier5.Cmp(stakemodifier6) != 0 {
		t.Error("The blockIDs are artificially generated and are the same every 256 blocks -> The stakemodifier is also the same ")
	}

}
*/
