package proxy

import (
	"encoding/binary"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/decred/dcrd/blockchain"
	"github.com/decred/dcrd/chaincfg/chainhash"
)

type BlockTemplate struct {
	Header     string
	Target     string
	Difficulty *big.Int
	Height     uint64
}

type Block struct {
	difficulty *big.Int
	header     string
}

func (b Block) Difficulty() *big.Int     { return b.difficulty }

func (b Block) Verify() bool {
	databytes, err := hex.DecodeString(b.header[0:360])
	if err != nil {
		return false
	}
	hash := chainhash.HashH(databytes)
	hashNum := blockchain.HashToBig(&hash)
	return hashNum.Cmp(b.difficulty) <= 0
}

func (s *ProxyServer) fetchBlockTemplate() {
	rpc := s.rpc()
	reply, err := rpc.GetWork()
	if err != nil {
		log.Printf("Error while refreshing block template on %s: %s", rpc.Name, err)
		return
	}

	t := s.currentBlockTemplate()

	height, diff, err := s.fetchPendingBlock(reply.Data)
	if err != nil {
		log.Printf("Error while refreshing pending block on %s: %s", rpc.Name, err)
		return
	}
	newTemplate := BlockTemplate{
		Header:     reply.Data,
		Target:     reply.Target,
		Height:     height,
		Difficulty: diff,
	}
	s.blockTemplate.Store(&newTemplate)

	if height != t.Height {
		log.Printf("New block to mine on %s at height: %d", rpc.Name, height)
	}
	return
}

func (s *ProxyServer) fetchPendingBlock(data string) (uint64, *big.Int, error) {
	blockNumberStr, err := hex.DecodeString(data[256:264])
	if err != nil {
		log.Println("Can't parse pending block number")
		// TODO: valami alap difficultyt!!
		return 0, nil, err
	}
	blockNumber := binary.LittleEndian.Uint32(blockNumberStr)
	blockDiffStr, err := hex.DecodeString(data[232:240])
	if err != nil {
		log.Println("Can't parse pending block difficulty")
		return 0, nil, err
	}
	blockDiff := binary.LittleEndian.Uint32(blockDiffStr)
	return uint64(blockNumber),  blockchain.CompactToBig(blockDiff), nil
}
