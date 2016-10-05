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
	//"io"
)

var BLOCKDELAY = 2000 //to calculate the stakemodifier
var PERIOD = 60       //period time in seconds on average between two blocks

type hash [32]byte

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
	hashlist := CalcGenesisBlocks()
	bsparameters.UpdateStakeModifier(hashlist)

	timestamp := time.Now().Unix()

	const NUMOFBCN = 20

	var Wallets [NUMOFBCN]wallet

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	var TotalBS uint64
	numof := 0

	for j := 0; j < NUMOFBCN; j++ {
		numin := r.Intn(5) + 1
		for i := 0; i < numin; i++ {
			newUTXO := UTXO{amount: uint64(r.Intn(100)), index: uint32(numof), blockHeight: uint32(0)}
			Wallets[j].UTXOs = append(Wallets[j].UTXOs, newUTXO)
			numof++
			TotalBS += Wallets[j].UTXOs[i].amount
		}
	}
	fmt.Println("TotalBS", TotalBS)
	bsparameters.UpdateDifficulty(PERIOD, TotalBS)

	counter := 0
	blockhigh := 0

	for i := 0; i < (3600 * 24); i++ {
		blockfound := false
		for j := 0; j < NUMOFBCN; j++ {
			winner := Wallets[j].CheckWinner(timestamp+int64(i), bsparameters)
			if winner >= 0 {
				//fmt.Println("BCN", j, "won with UTXO :", winner)
				Wallets[j].UTXOs[winner].index = 3
				Wallets[j].UTXOs[winner].blockHeight = uint32(blockhigh)
				blockhigh++
				blockfound = true
			}
		}
		if blockfound {
			//fmt.Println("block found in second: ", i)
			counter++
		}
	}
	fmt.Println("num of blocks found", counter, "not unique", blockhigh)
}

//calculate the stakemodifier out of the hashlist. automatically takes the last block as reference.
func (b *BSParameters) UpdateStakeModifier(hashlist []hash) {
	if len(hashlist) < (BLOCKDELAY + 255) {
		panic("hashlist is not long enough")
	}
	j := len(hashlist) - BLOCKDELAY
	for i := 0; i < 256; i++ {
		if i%8 == 0 {
			b.stakemodifier[i/8] = 0
		}
		b.stakemodifier[i/8] += hashlist[j-i][i/8] & (1 << (uint)(i%8))
	}
}

func (b *BSParameters) UpdateDifficulty(period int, totalbs uint64 /*blockchain interval*/) {
	b.difficulty = *new(big.Int).Div(new(big.Int).Div(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(int64(period))), big.NewInt(int64(totalbs)))

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
func CalcGenesisBlocks() []hash {
	var hashlist []hash
	for block := -(BLOCKDELAY + 255); block < 1; block++ {
		var buffer bytes.Buffer
		buffer.WriteString("genesis" + strconv.Itoa(block))
		hashlist = append(hashlist, sha256.Sum256(buffer.Bytes()))
	}
	return hashlist
}
