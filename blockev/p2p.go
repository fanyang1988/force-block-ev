package blockev

import (
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fanyang1988/force-block-ev/log"

	eos "github.com/eosforce/goforceio"
	"github.com/eosforce/goforceio/p2p"
	"go.uber.org/zap"
)

type Envelope struct {
	Peer    string     `json:"peer"`
	Packet  eos.Packet `json:"packet"`
	IsClose bool
}

type Handler interface {
	Handle(envelope *Envelope)
}

// P2PPeers a manager for peers to diff p2p node
type P2PPeers struct {
	name      string
	clients   []*p2p.Client
	handlers  []Handler
	msgChan   chan Envelope
	wg        sync.WaitGroup
	chanWg    sync.WaitGroup
	hasClosed bool
	mutex     sync.RWMutex
	logger    *zap.Logger
}

// NewP2PPeers new p2p peers from cfg
func NewP2PPeers(name string, chainID string, startBlock *eos.BlockHeader, peers []string) *P2PPeers {
	p := &P2PPeers{
		name:     name,
		clients:  make([]*p2p.Client, 0, len(peers)),
		handlers: make([]Handler, 0, 8),
		msgChan:  make(chan Envelope, 64),
		logger:   log.Logger(),
	}

	cID, err := hex.DecodeString(chainID)
	if err != nil {
		p.logger.Error("decode chain id err", zap.Error(err))
		panic(err)
	}

	var startBlockNum uint32 = 1
	var startBlockId eos.Checksum256
	var startBlockTime time.Time
	if startBlock != nil {
		startBlockId, _ = startBlock.BlockID()
		startBlockNum = startBlock.BlockNumber()
		startBlockTime = startBlock.Timestamp.Time
	}
	for idx, peer := range peers {
		p.logger.Debug("new peer client", zap.Int("idx", idx), zap.String("peer", peer))
		client := p2p.NewClient(
			p2p.NewOutgoingPeer(peer, fmt.Sprintf("%s-%02d", name, idx), &p2p.HandshakeInfo{
				ChainID:       cID,
				HeadBlockNum:  startBlockNum,
				HeadBlockID:   startBlockId,
				HeadBlockTime: startBlockTime,
			}),
			true,
		)
		client.RegisterHandler(p)
		p.clients = append(p.clients, client)
	}

	return p
}

// RegisterHandler register handler to p2p peers
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
				p.logger.Info("p2p peers stop")
				return
			}
		}
	}()

	for idx, client := range p.clients {
		p.createClient(idx, client)
	}
}

func (p *P2PPeers) isClosed() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.hasClosed
}

func (p *P2PPeers) createClient(idx int, client *p2p.Client) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			p.logger.Info("create connect", zap.Int("client", idx))
			err := client.Start()

			// check when after close client
			if p.isClosed() {
				return
			}

			if err != nil {
				p.logger.Error("client err", zap.Int("client", idx), zap.Error(err))
			}

			time.Sleep(3 * time.Second)

			// check when after sleep
			if p.isClosed() {
				return
			}
		}
	}()
}

func (p *P2PPeers) Close() {
	p.logger.Warn("start close")

	p.mutex.Lock()
	p.hasClosed = true
	p.mutex.Unlock()

	for idx, client := range p.clients {
		go func(i int, cli *p2p.Client) {
			err := cli.CloseConnection()
			if err != nil {
				p.logger.Error("client close err", zap.Int("client", i), zap.Error(err))
			}
			p.logger.Info("client close", zap.Int("client", i))
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
		p.logger.Warn("p2p peers msg chan closed")
		return true
	}

	for _, h := range p.handlers {
		func(hh Handler) {
			defer func() {
				if err := recover(); err != nil {
					p.logger.Error("handler process ev panic",
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
