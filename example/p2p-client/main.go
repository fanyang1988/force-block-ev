package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	eos "github.com/eosforce/goforceio"
	"github.com/eosforce/goforceio/p2p"
	"github.com/fanyang1988/force-block-ev/blockev"
)

var chainID = flag.String("chain-id", "bd61ae3a031e8ef2f97ee3b0e62776d6d30d4833c8f7c1645c657b149151004b", "net chainID to connect to")
var showLog = flag.Bool("v", false, "show detail log")
var startNum = flag.Int("num", 1, "start block num to sync")

// Wait wait for term signal, then stop the server
func Wait() {
	stopSignalChan := make(chan os.Signal, 1)
	signal.Notify(stopSignalChan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGUSR1)
	<-stopSignalChan
}

type handlerImp struct {
}

func (h *handlerImp) OnBlock(peer string, msg *eos.SignedBlock) error {
	return nil
}
func (h *handlerImp) OnGoAway(peer string, msg *eos.GoAwayMessage) error {
	return nil
}
func (h *handlerImp) OnHandshake(peer string, msg *eos.HandshakeMessage) error {
	return nil
}
func (h *handlerImp) OnTimeMsg(peer string, msg *eos.TimeMessage) error {
	return nil
}

func main() {
	flag.Parse()

	if *showLog {
		p2p.EnableP2PLogging()
		blockev.EnableLogging()

	}
	defer p2p.SyncLogger()

	// from 9001 - 9020
	const maxNumListen int = 20

	peers := make([]string, 0, maxNumListen+1)

	for i := 0; i < maxNumListen; i++ {
		peers = append(peers, fmt.Sprintf("127.0.0.1:%d", 8101+i))
	}
	peers = append(peers, "127.0.0.1:9999")

	p2pPeers := blockev.NewP2PPeers("test", *chainID, uint32(*startNum), peers)
	p2pPeers.RegisterHandler(blockev.LoggerHandler{})
	p2pPeers.RegisterHandler(blockev.NewP2PMsgHandler(&handlerImp{}))
	p2pPeers.Start()

	Wait()

	p2pPeers.Close()
}
