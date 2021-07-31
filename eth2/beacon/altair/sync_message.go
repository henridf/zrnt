package altair

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type SyncCommitteeMessage struct {
	// Slot to which this contribution pertains
	Slot common.Slot `yaml:"slot" json:"slot"`
	// Block root for this signature
	BeaconBlockRoot common.Root `yaml:"beacon_block_root" json:"beacon_block_root"`
	// Index of the validator that produced this signature
	ValidatorIndex common.ValidatorIndex `yaml:"validator_index" json:"validator_index"`
	// Signature by the validator over the block root of `slot`
	Signature common.BLSSignature `yaml:"signature" json:"signature"`
}

var SyncCommitteeMessageType = ContainerType("SyncCommitteeMessage", []FieldDef{
	{"slot", common.SlotType},
	{"beacon_block_root", RootType},
	{"validator_index", common.ValidatorIndexType},
	{"signature", common.BLSSignatureType},
})

func (agg *SyncCommitteeMessage) Deserialize(dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(
		&agg.Slot,
		&agg.BeaconBlockRoot,
		&agg.ValidatorIndex,
		&agg.Signature,
	)
}

func (agg *SyncCommitteeMessage) Serialize(w *codec.EncodingWriter) error {
	return w.FixedLenContainer(
		&agg.Slot,
		&agg.BeaconBlockRoot,
		&agg.ValidatorIndex,
		&agg.Signature,
	)
}

func (agg *SyncCommitteeMessage) ByteLength() uint64 {
	return codec.ContainerLength(
		&agg.Slot,
		&agg.BeaconBlockRoot,
		&agg.ValidatorIndex,
		&agg.Signature,
	)
}

func (agg *SyncCommitteeMessage) FixedLength() uint64 {
	return codec.ContainerLength(
		&agg.Slot,
		&agg.BeaconBlockRoot,
		&agg.ValidatorIndex,
		&agg.Signature,
	)
}

func (agg *SyncCommitteeMessage) HashTreeRoot(hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(
		&agg.Slot,
		&agg.BeaconBlockRoot,
		&agg.ValidatorIndex,
		&agg.Signature,
	)
}

type SyncCommitteeMessageView struct {
	*ContainerView
}

func AsSyncCommitteeMessage(v View, err error) (*SyncCommitteeMessageView, error) {
	c, err := AsContainer(v, err)
	return &SyncCommitteeMessageView{c}, err
}
