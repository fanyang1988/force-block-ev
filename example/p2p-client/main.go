package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/eosforce/goforceio/p2p"
)

var chainID = flag.String("chain-id", "bd61ae3a031e8ef2f97ee3b0e62776d6d30d4833c8f7c1645c657b149151004b", "net chainID to connect to")
var showLog = flag.Bool("v", false, "show detail log")

func main() {
	flag.Parse()

	if *showLog {
		p2p.EnableP2PLogging()
	}
	defer p2p.SyncLogger()

	cID, err := hex.DecodeString(*chainID)
	if err != nil {
		log.Fatal(err)
	}

	// from 9001 - 9020
	const maxNumListen int = 20

	for i := 0; i < maxNumListen; i++ {
		go func(idx int) {
			peer := fmt.Sprintf("127.0.0.1:%d", 9001+idx)
			fmt.Println("P2P Client ", peer, " With Chain ID :", *chainID)
			client := p2p.NewClient(
				p2p.NewOutgoingPeer(peer, fmt.Sprintf("eos-proxy-%02d", idx), &p2p.HandshakeInfo{
					ChainID:      cID,
					HeadBlockNum: 1,
				}),
				true,
			)

			client.RegisterHandler(p2p.StringLoggerHandler)
			client.Start()
		}(i)
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
