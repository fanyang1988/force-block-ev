package blockdb

import (
	eos "github.com/eosforce/goforceio"
	"github.com/pkg/errors"
)

// a Just simple imp

type BlockDB struct {
	peers map[string]*PeerBlockState
}

func (b *BlockDB) Init(peers []string) {
	b.peers = make(map[string]*PeerBlockState, len(peers))
	for _, peer := range peers {
		n := &PeerBlockState{
			peer: peer,
		}
		n.init()
		b.peers[peer] = n
	}
}

func (b *BlockDB) OnBlock(peer string, block *eos.SignedBlock) error {
	ps, ok := b.peers[peer]
	if !ok || ps == nil {
		return errors.New("no peer")
	}

	return ps.appendBlock(block)
}
