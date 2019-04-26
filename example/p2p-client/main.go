package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/zap"

	"github.com/fanyang1988/force-block-ev/log"

	"github.com/fanyang1988/force-block-ev/blockdb"

	eos "github.com/eosforce/goforceio"
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
	verifier *blockdb.FastBlockVerifier
}

func (h *handlerImp) OnBlock(peer string, msg *eos.SignedBlock) error {
	return h.verifier.OnBlock(peer, msg)
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

type verifyHandlerImp struct {
}

func (h *verifyHandlerImp) OnBlock(blockNum uint32, blockID eos.Checksum256, block *eos.SignedBlock) error {
	log.Logger().Info("on checked block",
		zap.Uint32("num", blockNum), zap.String("id", blockID.String()), zap.Int("trx num", len(block.Transactions)))
	return nil
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(8)

	if *showLog {
		log.EnableLogging(false)
	}

	// from 9001 - 9020
	const maxNumListen int = 20

	peers := make([]string, 0, maxNumListen+1)

	for i := 0; i < maxNumListen; i++ {
		peers = append(peers, fmt.Sprintf("127.0.0.1:%d", 8101+i))
	}
	peers = append(peers, "127.0.0.1:9999")

	p2pPeers := blockev.NewP2PPeers("test", *chainID, nil, peers)
	//p2pPeers.RegisterHandler(blockev.LoggerHandler{})
	p2pPeers.RegisterHandler(blockev.NewP2PMsgHandler(&handlerImp{
		verifier: blockdb.NewFastBlockVerifier(peers, &verifyHandlerImp{}),
	}))
	p2pPeers.Start()

	Wait()

	p2pPeers.Close()
}
