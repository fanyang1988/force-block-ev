package blockev

import (
	eos "github.com/eosforce/goforceio"
	"go.uber.org/zap"
)

type P2PMsgHandler struct {
}

// Handle imp Handler func
func (m P2PMsgHandler) Handle(envelope *Envelope) {
	if envelope.IsClose == true {
		return
	}

	var err error
	switch envelope.Packet.Type {
	case eos.GoAwayMessageType:
		err = m.handlerGoAway(envelope.Peer, envelope.Packet.P2PMessage.(*eos.GoAwayMessage))
	case eos.SignedBlockType:
		err = m.handlerBlock(envelope.Peer, envelope.Packet.P2PMessage.(*eos.SignedBlock))
	}

	if err != nil {
		logger.Error("handle msg err", zap.Error(err))
	}
}

func (m P2PMsgHandler) handlerBlock(peer string, block *eos.SignedBlock) error {
	id, err := block.BlockID()
	if err != nil {
		return err
	}
	logger.Debug("on block",
		zap.String("peer", peer),
		zap.Uint32("num", block.BlockNumber()), zap.String("id", id.String()))
	return nil
}

func (m P2PMsgHandler) handlerGoAway(peer string, goAway *eos.GoAwayMessage) error {
	logger.Debug("go away",
		zap.String("peer", peer), zap.String("reason", goAway.Reason.String()))
	return nil
}
