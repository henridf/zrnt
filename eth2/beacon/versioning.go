package beacon

import (
	. "github.com/protolambda/ztyp/view"
)

const VersionType = Bytes4Type

// 32 bits, not strictly an integer, hence represented as 4 bytes
// (bytes not necessarily corresponding to versions)
type Version [4]byte

func (v Version) ToUint32() uint32 {
	return uint32(v[0])<<24 | uint32(v[1])<<16 | uint32(v[2])<<8 | uint32(v[3])
}

func (v Version) View() SmallByteVecView {
	return v[:]
}

func AsVersion(v View, err error) (Version, error) {
	return AsBytes4(v, err)
}

type Fork struct {
	PreviousVersion Version
	CurrentVersion Version
	Epoch Epoch
}

func (f *Fork) View() *ForkView {
	res, _ := CheckpointType.FromFields(f.PreviousVersion.View(), f.CurrentVersion.View(), Uint64View(f.Epoch))
	return &ForkView{res}
}

var ForkType = ContainerType("Fork", []FieldDef{
	{"previous_version", VersionType},
	{"current_version", VersionType},
	{"epoch", EpochType}, // Epoch of latest fork
})

type ForkView struct { *ContainerView }

func (f *ForkView) PreviousVersion() (Version, error) {
	return AsVersion(f.Get(0))
}

func (f *ForkView) CurrentVersion() (Version, error) {
	return AsVersion(f.Get(1))
}

func (f *ForkView) Epoch() (Epoch, error) {
	return AsEpoch(f.Get(2))
}

func (f *ForkView) Raw() (*Fork, error) {
	prev, err := f.PreviousVersion()
	if err != nil {
		return nil, err
	}
	curr, err := f.CurrentVersion()
	if err != nil {
		return nil, err
	}
	ep, err := f.Epoch()
	if err != nil {
		return nil, err
	}
	return &Fork{
		PreviousVersion: prev,
		CurrentVersion:  curr,
		Epoch:           ep,
	}, nil
}

func AsFork(v View, err error) (*ForkView, error) {
	c, err := AsContainer(v, err)
	return &ForkView{c}, err
}

// Return the signature domain (fork version concatenated with domain type) of a message.
func (state *BeaconStateView) GetDomain(dom BLSDomainType, messageEpoch Epoch) (BLSDomain, error) {
	forkView, err := state.Fork()
	if err != nil {
		return BLSDomain{}, err
	}
	fork, err := forkView.Raw()
	if err != nil {
		return BLSDomain{}, err
	}
	var v Version
	if messageEpoch < fork.Epoch {
		v = fork.PreviousVersion
	} else {
		v = fork.CurrentVersion
	}
	genesisValRoot, err := state.GenesisValidatorsRoot()
	if err != nil {
		return BLSDomain{}, err
	}
	// combine fork version with domain type.
	return ComputeDomain(dom, v, genesisValRoot), nil
}