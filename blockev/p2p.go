package blockev

import (
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"sync"

	eos "github.com/eosforce/goforceio"
	"github.com/eosforce/goforceio/p2p"
	"go.uber.org/zap"
)

var logger = zap.NewNop()

func EnableLogging() {
	logger = eos.NewLogger(false)
}

type Envelope struct {
	Peer    string     `json:"peer"`
	Packet  eos.Packet `json:"packet"`
	IsClose bool
}

type Handler interface {
	Handle(envelope *Envelope)
}

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

// P2PPeers a manager for peers to diff p2p node
type P2PPeers struct {
	name     string
	clients  []*p2p.Client
	handlers []Handler
	msgChan  chan Envelope
	wg       sync.WaitGroup
	chanWg   sync.WaitGroup
}

// NewP2PPeers new p2p peers from cfg
func NewP2PPeers(name string, chainID string, startBlockNum uint32, peers []string) *P2PPeers {
	p := &P2PPeers{
		name:     name,
		clients:  make([]*p2p.Client, 0, len(peers)),
		handlers: make([]Handler, 0, 8),
		msgChan:  make(chan Envelope, 64),
	}

	cID, err := hex.DecodeString(chainID)
	if err != nil {
		logger.Error("decode chain id err", zap.Error(err))
		panic(err)
	}

	for idx, peer := range peers {
		logger.Debug("new peer client", zap.Int("idx", idx), zap.String("peer", peer))
		client := p2p.NewClient(
			p2p.NewOutgoingPeer(peer, fmt.Sprintf("%s-%02d", name, idx), &p2p.HandshakeInfo{
				ChainID:      cID,
				HeadBlockNum: startBlockNum,
			}),
			true,
		)
		client.RegisterHandler(p)
		p.clients = append(p.clients, client)
	}

	return p
}

// RegisterHandler register handler to p2ppeers
func (p *P2PPeers) RegisterHandler(handler Handler) {
	p.handlers = append(p.handlers, handler)
}

func (p *P2PPeers) Start() {
	p.chanWg.Add(1)
	go func() {
		defer p.chanWg.Done()
		for {
			isStop := p.Loop()
			if isStop {
				logger.Info("p2p peers stop")
				return
			}
		}
	}()

	for idx, client := range p.clients {
		p.wg.Add(1)
		go func(i int, cli *p2p.Client) {
			defer p.wg.Done()
			// TODO close
			err := cli.Start()
			if err != nil {
				logger.Error("client err", zap.Int("client", i), zap.Error(err))
			}
		}(idx, client)
	}
}

func (p *P2PPeers) Close() {
	logger.Warn("start close")
	for idx, client := range p.clients {
		go func(i int, cli *p2p.Client) {
			cli.CloseConnection()
			logger.Info("client close", zap.Int("client", i))
		}(idx, client)
	}
	p.wg.Wait()
	p.msgChan <- Envelope{
		IsClose: true,
	}
	close(p.msgChan)
	p.chanWg.Wait()
}

func (p *P2PPeers) Loop() bool {
	ev, ok := <-p.msgChan
	if ev.IsClose {
		return true
	}

	if !ok {
		logger.Warn("p2p peers msg chan closed")
		return true
	}

	for _, h := range p.handlers {
		func(hh Handler) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("handler process ev panic",
						zap.String("err", fmt.Sprintf("err:%s", err)),
						zap.String("stack", string(debug.Stack())))
				}
			}()
			hh.Handle(&ev)
		}(h)
	}

	return false
}

// Handle handler for p2p clients
func (p *P2PPeers) Handle(envelope *p2p.Envelope) {
	p.msgChan <- Envelope{
		Peer:   envelope.Sender.Address,
		Packet: *envelope.Packet,
	}
}
