package blockev

type HandlerFunc func(envelope *Envelope)

func (f HandlerFunc) Handle(envelope *Envelope) {
	f(envelope)
}

type LoggerHandler struct {
}

func (f LoggerHandler) Handle(envelope *Envelope) {
	logger.Sugar().Infof("handler ev from %s by %s",
		envelope.Peer, envelope.Packet.P2PMessage.String())
}
