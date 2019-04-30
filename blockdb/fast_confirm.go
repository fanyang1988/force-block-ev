package blockdb

import (
	eos "github.com/eosforce/goforceio"
	"github.com/fanyang1988/force-block-ev/log"
	"go.uber.org/zap"
)

type VerifyHandler interface {
	OnBlock(blockNum uint32, blockID eos.Checksum256, block *eos.SignedBlock) error
}

type FastBlockVerifier struct {
	db            *BlockDB
	verifyHandler VerifyHandler
	lastVerify    blockItem
	startBlock    uint32
}

// NewFastBlockVerifier create
func NewFastBlockVerifier(peers []string, startBlock uint32, verifyHandler VerifyHandler) *FastBlockVerifier {
	blocks := &BlockDB{}
	blocks.Init(peers)
	res := &FastBlockVerifier{
		db:            blocks,
		verifyHandler: verifyHandler,
		startBlock:    startBlock,
	}
	if startBlock > 1 {
		res.lastVerify.blockNum = startBlock - 1
	}
	return res
}

// CheckBlocks Call when append block to db
func (f *FastBlockVerifier) checkBlocks() error {
	bi, ok := f.TryGetVerifyBlock()
	if ok {
		err := f.verifyHandler.OnBlock(bi.blockNum, bi.blockID, bi.block)
		if err != nil {
			// no del block
			return err
		}
		f.lastVerify = *bi
		f.db.DelBlockBefore(f.lastVerify.blockNum - 3)

		return nil
	}

	return nil
}

// OnBlock
func (f *FastBlockVerifier) OnBlock(peer string, block *eos.SignedBlock) error {
	if block.BlockNumber() > f.lastVerify.blockNum {
		err := f.db.OnBlock(peer, block)
		if err != nil {
			return err
		}
	}
	return f.checkBlocks()
}

func (f *FastBlockVerifier) TryGetVerifyBlock() (*blockItem, bool) {
	// TODO now is a simple imp
	num2Verify := f.lastVerify.blockNum + 1
	var bi *blockItem
	for peer, stat := range f.db.Peers {
		// 1. every peer blocks are no fork
		for _, blocks := range stat.blocks {
			if len(blocks.blocks) > 1 {
				log.Logger().Info("peer block forking",
					zap.String("peer", peer),
					zap.Uint32("num", blocks.blockNum),
					zap.Int("fork", len(blocks.blocks)))
				return nil, false
			}
		}

		// 2. all peer are same
		first := stat.GetBlock(num2Verify)
		if bi == nil && first != nil {
			bi = first
		} else {
			if (first == nil) ||
				(!IsChecksum256Eq(bi.blockID, first.blockID)) ||
				(!IsChecksum256Eq(bi.block.Previous, first.block.Previous)) {
				return nil, false
			}
		}
	}

	if bi == nil {
		return nil, false
	}

	return bi, true
}
