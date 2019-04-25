package blockdb

import eos "github.com/eosforce/goforceio"

const defaultForkBlocksInOneNumSize int = 32

type blockItem struct {
	preBlockIdx int
	blockNum    uint32
	blockID     eos.Checksum256
	block       *eos.SignedBlock
}

type blockInNum struct {
	blockNum uint32
	blocks   []blockItem
}

func (b *blockInNum) find(block *eos.SignedBlock) int {
	for idx, bk := range b.blocks {
		if IsBlockEq(bk.block, block) {
			return idx
		}
	}
	return -1
}

func (b *blockInNum) findByID(blockID eos.Checksum256) int {
	for idx, bk := range b.blocks {
		if IsChecksum256Eq(blockID, bk.blockID) {
			return idx
		}
	}
	return -1
}

func (b *blockInNum) append(bi blockItem) int {
	idx := b.find(bi.block)
	if idx >= 0 {
		return idx
	}

	b.blocks = append(b.blocks, bi)
	return len(b.blocks) - 1
}

type PeerBlockState struct {
	peer   string
	blocks []blockInNum
}

func (p *PeerBlockState) init() {
	p.blocks = make([]blockInNum, 0, 512)
}

func (p *PeerBlockState) newBlockNum() uint32 {
	if len(p.blocks) == 0 {
		return 0
	}
	return p.blocks[len(p.blocks)-1].blockNum
}

func (p *PeerBlockState) appendBlock(block *eos.SignedBlock) error {
	// TODO just a simple imp
	blockNum := block.BlockNumber()
	blockID, err := block.BlockID()
	if err != nil {
		return err
	}
	if len(p.blocks) == 0 {
		// no init
		p.blocks = append(p.blocks, blockInNum{
			blockNum: blockNum,
			blocks:   make([]blockItem, 0, defaultForkBlocksInOneNumSize),
		})
	}

	newBlockNum := p.newBlockNum()
	firstBlockNum := p.blocks[0].blockNum

	if blockNum < firstBlockNum {
		// no need process
		return nil
	}

	if blockNum > (newBlockNum + 1) {
		// no need process
		return nil
	}

	if blockNum == (newBlockNum + 1) {
		for i := 0; i < int(blockNum-newBlockNum); i++ {
			p.blocks = append(p.blocks, blockInNum{
				blockNum: newBlockNum + uint32(i) + 1,
				blocks:   make([]blockItem, 0, defaultForkBlocksInOneNumSize),
			})
		}
	}

	var perBlockIdx int = -1
	if blockNum > firstBlockNum {
		perBlockItemIdx := blockNum - firstBlockNum - 1
		perBlockIdx = p.blocks[perBlockItemIdx].findByID(blockID)
		if perBlockIdx < 0 {
			// no process no pre block id
			return nil
		}
	}

	p.blocks[blockNum-firstBlockNum].append(blockItem{
		perBlockIdx,
		blockNum,
		blockID,
		block,
	})

	return nil
}
