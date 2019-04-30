package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	eos "github.com/eosforce/goforceio"
	"github.com/eosforce/goforceio/p2p"
	"github.com/fanyang1988/force-block-ev/blockdb"
	"github.com/fanyang1988/force-block-ev/blockev"
	"github.com/fanyang1988/force-block-ev/log"
	"github.com/fanyang1988/force-go"
	"github.com/fanyang1988/force-go/config"
	"github.com/fanyang1988/force-go/types"
	"go.uber.org/zap"
)

var chainID = flag.String("chain-id", "bd61ae3a031e8ef2f97ee3b0e62776d6d30d4833c8f7c1645c657b149151004b", "net chainID to connect to")
var showLog = flag.Bool("v", false, "show detail log")
var startNum = flag.Int("num", 1, "start block num to sync")
var p2pAddress = flag.String("p2p", "", "p2p address")
var url = flag.String("url", "", "p2p address")

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

func (h *handlerImp) OnBlock(peer string, msg *types.BlockGeneralInfo) error {
	log.Logger().Info("on checked block")
	return h.verifier.OnBlock(peer, msg)
}
func (h *handlerImp) OnGoAway(peer string, reason uint8, nodeID types.Checksum256) error {
	return nil
}

type verifyHandlerImp struct {
}

func (h *verifyHandlerImp) OnBlock(block *types.BlockGeneralInfo) error {
	log.Logger().Info("on checked block",
		zap.Uint32("num", block.BlockNum), zap.String("id", block.ID.String()), zap.Int("trx num", len(block.Transactions)))
	return nil
}

func getBlockBegin(num uint32) *types.BlockGeneralInfo {
	client, err := force.NewClient(types.FORCEIO, &config.ConfigData{
		ChainID: *chainID,
		URL:     *url,
	})
	if err != nil {
		panic(err)
	}

	b, err := client.GetBlockDataByNum(num)
	if err != nil {
		panic(err)
	}

	log.Logger().Sugar().Infof("get block %v %v %v", b.BlockNum, b.ID, *b)

	return b
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(8)

	if *showLog {
		log.EnableLogging(false)
		p2p.EnableP2PLogging()
	}

	// from 9001 - 9020
	const maxNumListen int = 20
	peers := make([]string, 0, maxNumListen+1)

	if *p2pAddress == "" {
		for i := 0; i < maxNumListen; i++ {
			peers = append(peers, fmt.Sprintf("127.0.0.1:%d", 8101+i))
		}
		peers = append(peers, "127.0.0.1:9999")
	} else {
		peers = append(peers, *p2pAddress)
	}

	var stratBlock *blockev.P2PSyncData
	if *startNum != 0 {
		block := getBlockBegin(uint32(*startNum))
		blockirr := getBlockBegin(uint32(*startNum) - 15)

		stratBlock = &blockev.P2PSyncData{
			HeadBlockNum:             block.BlockNum,
			HeadBlockID:              eos.Checksum256(block.ID),
			HeadBlockTime:            block.Timestamp,
			LastIrreversibleBlockNum: blockirr.BlockNum,
			LastIrreversibleBlockID:  eos.Checksum256(blockirr.ID),
		}
		log.Logger().Sugar().Infof("start %v", *stratBlock)
	}

	p2pPeers := blockev.NewP2PPeers("test", *chainID, stratBlock, peers)
	p2pPeers.RegisterHandler(blockev.LoggerHandler{})
	log.Logger().Info("dd")
	p2pPeers.RegisterHandler(blockev.NewP2PMsgHandler(&handlerImp{
		verifier: blockdb.NewFastBlockVerifier(peers, uint32(*startNum), &verifyHandlerImp{}),
	}))
	p2pPeers.Start()

	Wait()

	p2pPeers.Close()
}
