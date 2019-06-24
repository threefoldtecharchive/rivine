package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	"github.com/threefoldtech/rivine/build"
	//"io"
)

var BLOCKDELAY = 2000  //to calculate the stakemodifier
var PERIOD = 60        //period time in seconds on average between two blocks
var DIFF_REFRESH = 500 //number of blocks period to refresh before refresh the diff
var DIFF_STRENGTH = 500

type hash [32]byte

type blockchainitem struct {
	blockhash hash
	timestamp int64
}

type UTXO struct {
	blockHeight uint32
	index       uint32
	amount      uint64
}

type wallet struct {
	UTXOs []UTXO
}

type BSParameters struct {
	difficulty    big.Int
	stakemodifier hash
}

func main() {
	//test the functions
	var bsparameters BSParameters

	timestamp := time.Now().Unix()
	blockchain := CalcGenesisBlocks(timestamp)
	bsparameters.UpdateStakeModifier(blockchain)
	const NUMOFBCN = 20

	var Wallets [NUMOFBCN]wallet

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	NumOfHashes := new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)

	var TotalBS uint64
	numof := 0

	for j := 0; j < NUMOFBCN; j++ {
		numin := r.Intn(5) + 1
		for i := 0; i < numin; i++ {
			newUTXO := UTXO{amount: uint64(r.Intn(100) * 100), index: uint32(numof), blockHeight: uint32(0)}
			Wallets[j].UTXOs = append(Wallets[j].UTXOs, newUTXO)
			numof++
			TotalBS += Wallets[j].UTXOs[i].amount
		}
	}
	fmt.Println("TotalBS", TotalBS)
	StartDifficulty := *new(big.Int).Div(new(big.Int).Div(NumOfHashes, big.NewInt(int64(PERIOD))), big.NewInt(int64(TotalBS)))
	bsparameters.difficulty = StartDifficulty

	counter := 0
	blockhigh := 0

	for i := 0; i < (3600 * 24 * 300); i++ {
		blockfound := false
		numofbcns := NUMOFBCN
		if i > (120000) {
			//	numofbcns -= 5
		}
		if i > (500000) {
			//	numofbcns -= 10
		}
		TotalBS = 0
		for j := 0; j < numofbcns; j++ {
			winner := Wallets[j].CheckWinner(timestamp+int64(i), bsparameters)
			for k := 0; k < len(Wallets[j].UTXOs); k++ {
				TotalBS += Wallets[j].UTXOs[k].amount
			}
			if winner >= 0 {
				//fmt.Println("BCN", j, "won with UTXO :", winner)
				Wallets[j].UTXOs[winner].index = 3
				Wallets[j].UTXOs[winner].blockHeight = uint32(blockhigh)
				blockhigh++
				blockfound = true
			}
		}
		if blockfound {
			var buffer bytes.Buffer
			buffer.WriteString("random" + strconv.Itoa(i))
			blockchain = append(blockchain, blockchainitem{sha256.Sum256(buffer.Bytes()), timestamp + int64(i)})
			bsparameters.UpdateStakeModifier(blockchain)
			//fmt.Println("block found in second: ", i)
			counter++
			if counter > DIFF_STRENGTH && counter%DIFF_REFRESH == 0 {
				bsparameters.UpdateDifficulty(NumOfHashes, PERIOD, DIFF_STRENGTH, blockchain)
				fmt.Println(counter, ",", i/PERIOD, ",", i/PERIOD-counter, ",", TotalBS)
			}
		}
	}
	fmt.Println("num of blocks found", counter, "not unique", blockhigh)
}

//calculate the stakemodifier out of the blockchain. automatically takes the last block as reference.
func (b *BSParameters) UpdateStakeModifier(blockchain []blockchainitem) {
	if len(blockchain) < (BLOCKDELAY + 255) {
		build.Critical("blockchain is not long enough")
	}
	j := len(blockchain) - BLOCKDELAY
	for i := 0; i < 256; i++ {
		if i%8 == 0 {
			b.stakemodifier[i/8] = 0
		}
		b.stakemodifier[i/8] += blockchain[j-i].blockhash[i/8] & (1 << (uint)(i%8))
	}
}
func (b *BSParameters) UpdateDifficulty(NumOfHashes *big.Int, period int, diffstrenght int, blockchain []blockchainitem) {
	if len(blockchain) < (diffstrenght) {
		build.Critical("blockchain is not long enough")
	}
	diff := blockchain[len(blockchain)-1].timestamp - blockchain[len(blockchain)-diffstrenght-1].timestamp

	totalBScalc := *new(big.Int).Div(new(big.Int).Div(new(big.Int).Mul(NumOfHashes, big.NewInt(int64(diffstrenght))), big.NewInt(int64(diff))), &b.difficulty)

	b.difficulty = *new(big.Int).Div(NumOfHashes, new(big.Int).Mul(&totalBScalc, big.NewInt(int64(period))))

	Ratio := big.NewRat(diff, int64(period*diffstrenght))

	//	b.difficulty = *new(big.Int).Div(new(big.Int).Mul(&b.difficulty, big.NewInt(diff)), new(big.Int).Mul(big.NewInt(int64(period)), big.NewInt(int64(diffstrenght))))

	//	totalBScalc := *new(big.Int).Div(new(big.Int).Div(        , big.NewInt(int64(period))), &b.difficulty)

	fmt.Print(totalBScalc.String())
	fmt.Print(",", Ratio.RatString(), ",", float64(diff)/float64(period)/float64(diffstrenght), ",")
}

//checks if this wallet has a winner for this timestamp
//returns -1 if no winner , otherwise winner is the index of the utxo who won with the biggest amount
func (w *wallet) CheckWinner(timestamp int64, bsparam BSParameters) int {
	var biggestamoutwon uint64
	var result int
	result = -1
	for index, utxo := range w.UTXOs {
		if BlockCreation(bsparam.stakemodifier, utxo.blockHeight, utxo.index, timestamp, utxo.amount, &bsparam.difficulty) {
			if biggestamoutwon < utxo.amount {
				result = index
				biggestamoutwon = utxo.amount
			}
		}
	}
	return result
}

//calculates the blockcreation criterium, returns true if the criterium is met.
func BlockCreation(stakemodifier hash, blocknumber uint32, indexutxo uint32, timestamp int64, stakeamout uint64, difficulty *big.Int) bool {
	if stakeamout == 0 {
		return false
	}
	k := new(bytes.Buffer)
	binary.Write(k, binary.LittleEndian, blocknumber)
	index := new(bytes.Buffer)
	binary.Write(index, binary.LittleEndian, indexutxo)
	ts := new(bytes.Buffer)
	binary.Write(ts, binary.LittleEndian, timestamp)
	dat := append(append(append(stakemodifier[:], k.Bytes()...), index.Bytes()...), ts.Bytes()...)
	hash := sha256.Sum256(dat)
	res := new(big.Int).Div(new(big.Int).SetBytes(hash[:]), big.NewInt(int64(stakeamout)))

	if res.Cmp(difficulty) < 0 {
		return true
	}
	return false
}

//calculates the hashes of the blocks before the genesis block in a deterministic way.
func CalcGenesisBlocks(timestamp int64) []blockchainitem {
	var blockchain []blockchainitem
	for block := -(BLOCKDELAY + 255); block < 1; block++ {
		var buffer bytes.Buffer
		buffer.WriteString("genesis" + strconv.Itoa(block))
		blockchain = append(blockchain, blockchainitem{sha256.Sum256(buffer.Bytes()), timestamp})
	}
	return blockchain
}
