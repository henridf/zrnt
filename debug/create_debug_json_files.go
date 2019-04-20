package main

import (
	"encoding/hex"
	"fmt"
	"github.com/protolambda/zrnt/debug/util/debug_format"
	. "github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/beacon/genesis"
	"github.com/protolambda/zrnt/eth2/beacon/transition"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/util/hash"
	"github.com/protolambda/zrnt/eth2/util/math"
	"github.com/protolambda/zrnt/eth2/util/merkle"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"math/rand"
	"os"
)

func main() {

	// RNG used to create simulated blocks
	rng := rand.New(rand.NewSource(0xDEADBEEF))

	hashHexStr := func(value interface{}) string {
		encoded := ssz.HashTreeRoot(value)
		out := make([]byte, hex.EncodedLen(32))
		hex.Encode(out, encoded[:])
		return string(out)
	}

	genesisTime := Timestamp(1222333444)
	genesisValidatorCount := 100

	privKeys := make([][32]byte, 0, genesisValidatorCount)
	deposits := make([]Deposit, 0, genesisValidatorCount)
	depRoots := make([]Root, 0, genesisValidatorCount)
	for i := 0; i < genesisValidatorCount; i++ {
		// create a random 32 byte private key. We're not using real crypto yet.
		privKey := [32]byte{}
		rng.Read(privKey[:])
		privKeys = append(privKeys, privKey)
		// simply derive pubkey and withdraw creds, not real thing yet
		pubKey := BLSPubkey{}
		h := hash.Hash(privKey[:])
		copy(pubKey[:], h)
		withdrawCreds := Root{}
		copy(withdrawCreds[:], hash.Hash(append(h[:], 1)))
		dep := Deposit{
			Proof: [DEPOSIT_CONTRACT_TREE_DEPTH]Root{},
			Index: DepositIndex(i),
			Data: DepositData{
				Pubkey:                pubKey,
				WithdrawalCredentials: withdrawCreds,
				Amount:    Gwei(100),
				ProofOfPossession:     BLSSignature{1, 2, 3}, // BLS not yet implemented
			},
		}
		depLeafHash := Root{}
		copy(depLeafHash[:], hash.Hash(dep.Data.Serialized()))
		deposits = append(deposits, dep)
		depRoots = append(depRoots, depLeafHash)
	}
	for i := 0; i < len(deposits); i++ {
		copy(deposits[i].Proof[:], merkle.ConstructProof(depRoots, uint64(i), uint8(DEPOSIT_CONTRACT_TREE_DEPTH)))
	}
	power2 := math.NextPowerOfTwo(uint64(len(depRoots)))
	depositsRoot := merkle.MerkleRoot(depRoots)
	// Now pad with zero branches to complete depth.
	buf := [64]byte{}
	for i := power2; i < (1 << DEPOSIT_CONTRACT_TREE_DEPTH); i <<= 1 {
		copy(buf[0:32], depositsRoot[:])
		depositsRoot = hash.Hash(buf[:])
	}

	eth1Data := Eth1Data{
		DepositRoot: depositsRoot,
		BlockHash:   Root{42}, // TODO eth1 simulation
	}
	genesisState := genesis.GetGenesisBeaconState(deposits, genesisTime, eth1Data)

	preState := genesisState

	for i := 0; i < 300; i++ {

		block, err := SimulateBlock(preState, rng)
		if err != nil {
			panic(err)
		}
		name := fmt.Sprintf("block_%d_%s", i, hashHexStr(block))
		// create the data, encode it, and write it to a file
		if err := writeDebugJson(name, block); err != nil {
			panic(err)
		}
		state, err := transition.StateTransition(preState, block, true)
		if err != nil {
			panic(err)
		}
		block.StateRoot = ssz.HashTreeRoot(state)
		name = fmt.Sprintf("state_%d_%s", i, hashHexStr(state))
		// create the data, encode it, and write it to a file
		if err := writeDebugJson(name, state); err != nil {
			panic(err)
		}
		preState = state
	}

}

func writeDebugJson(name string, data interface{}) error {
	encodedData, err := debug_format.MarshalJSON(data, "    ")
	if err != nil {
		fmt.Println("Failed to encode for", name)
		return err
	}
	f, err := os.Create(name + ".json")
	defer f.Close()
	if err != nil {
		return err
	}
	if _, err := f.Write(encodedData); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println("encoded and written file for ", name)
	return nil
}

func SimulateBlock(state *BeaconState, rng *rand.Rand) (*BeaconBlock, error) {
	prevHeader := state.LatestBlockHeader
	// stub state root
	prevHeader.StateRoot = ssz.HashTreeRoot(state)

	parentRoot := ssz.HashTreeRoot(prevHeader)
	block := &BeaconBlock{
		Slot:              state.Slot + 1 + Slot(rng.Intn(5)),
		PreviousBlockRoot: parentRoot,
		StateRoot:         Root{},
		Body: BeaconBlockBody{
			RandaoReveal: BLSSignature{4, 2},
			Eth1Data: Eth1Data{
				DepositRoot: Root{0, 1, 3},
				BlockHash:   Root{4, 5, 6},
			},
			// no transfers
			// TODO simulate transfers
		},
		Signature: BLSSignature{1, 2, 3}, // TODO implement BLS
	}
	// TODO: set randao reveal
	// TODO: include eth1 data
	// TODO: sign proposal
	postState, err := transition.StateTransition(state, block, false)
	if err != nil {
		return nil, err
	}
	block.StateRoot = ssz.HashTreeRoot(postState)
	return block, nil
}
