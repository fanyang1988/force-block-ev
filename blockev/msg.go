package blockev

import (
	eos "github.com/eosforce/goforceio"
	"go.uber.org/zap"
)

type P2PMsgHandlerImp interface {
	OnBlock(peer string, msg *eos.SignedBlock) error
	OnGoAway(peer string, msg *eos.GoAwayMessage) error
	OnHandshake(peer string, msg *eos.HandshakeMessage) error
	OnTimeMsg(peer string, msg *eos.TimeMessage) error
}

type P2PMsgHandler struct {
	imp P2PMsgHandlerImp
}

// NewP2PMsgHandler Create a handler
func NewP2PMsgHandler(imp P2PMsgHandlerImp) *P2PMsgHandler {
	return &P2PMsgHandler{
		imp: imp,
	}
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

func (m P2PMsgHandler) handlerBlock(peer string, msg *eos.SignedBlock) error {
	id, err := msg.BlockID()
	if err != nil {
		return err
	}
	logger.Debug("on block",
		zap.String("peer", peer),
		zap.Uint32("num", msg.BlockNumber()), zap.String("id", id.String()))
	if m.imp != nil {
		return m.imp.OnBlock(peer, msg)
	}
	return nil
}

func (m P2PMsgHandler) handlerGoAway(peer string, msg *eos.GoAwayMessage) error {
	logger.Debug("go away",
		zap.String("peer", peer), zap.String("reason", msg.Reason.String()))
	if m.imp != nil {
		return m.imp.OnGoAway(peer, msg)
	}
	return nil
}

func (m P2PMsgHandler) handlerTime(peer string, msg *eos.TimeMessage) error {
	logger.Debug("on time",
		zap.String("peer", peer), zap.String("time", msg.String()))
	if m.imp != nil {
		return m.imp.OnTimeMsg(peer, msg)
	}
	return nil
}

func (m P2PMsgHandler) handlerHandshake(peer string, msg *eos.HandshakeMessage) error {
	logger.Debug("on handshake",
		zap.String("peer", peer), zap.String("handshake", msg.String()))
	if m.imp != nil {
		return m.imp.OnHandshake(peer, msg)
	}
	return nil
}
