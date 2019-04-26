package blockev

import (
	"github.com/eosforce/goeosforce/ecc"
	"github.com/fanyang1988/force-block-ev/log"
)

func init() {
	ecc.PublicKeyPrefixCompat = "FOSC"
}

type HandlerFunc func(envelope *Envelope)

func (f HandlerFunc) Handle(envelope *Envelope) {
	f(envelope)
}

type LoggerHandler struct {
}

func (f LoggerHandler) Handle(envelope *Envelope) {
	log.Logger().Sugar().Infof("handler ev from %s by %s",
		envelope.Peer, envelope.Packet.P2PMessage.String())
}
