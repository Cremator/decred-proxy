package proxy

import (
	"../util"
	"log"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Miner struct {
	sync.RWMutex
	Id            string
	IP            string
	startedAt     int64
	lastBeat      int64
	validShares   uint64
	invalidShares uint64
	accepts       uint64
	rejects       uint64
	shares        map[int64]int64
}

func NewMiner(id, ip string) *Miner {
	miner := &Miner{Id: id, IP: ip, shares: make(map[int64]int64), startedAt: util.MakeTimestamp()}
	return miner
}

func (m *Miner) heartbeat() {
	now := util.MakeTimestamp()
	atomic.StoreInt64(&m.lastBeat, now)
}

func (m *Miner) getLastBeat() int64 {
	return atomic.LoadInt64(&m.lastBeat)
}

func (m *Miner) storeShare(diff int64) {
	now := util.MakeTimestamp()
	m.Lock()
	m.shares[now] += diff
	m.Unlock()
}

func (m *Miner) hashrate(hashrateWindow time.Duration) int64 {
	now := util.MakeTimestamp()
	totalShares := int64(0)
	window := int64(hashrateWindow / time.Millisecond)
	boundary := now - m.startedAt

	if boundary > window {
		boundary = window
	}

	m.Lock()
	for k, v := range m.shares {
		if k < now-86400000 {
			delete(m.shares, k)
		} else if k >= now-window {
			totalShares += v
		}
	}
	m.Unlock()
	return totalShares / boundary
}

func (m *Miner) processShare(s *ProxyServer, t *BlockTemplate, diff string, params []string) bool {
	paramsOrig := params[:]

	rpc := s.rpc()
	var shareDiff *big.Int

	if !rpc.Pool && diff != "" {
		minerDifficulty, err := strconv.ParseFloat(diff, 64)
		if err != nil {
			log.Println("Malformed difficulty: " + diff)
			minerDifficulty = 8
		}
		if (minerDifficulty <= 0) {
			log.Println("Invalid difficulty: " + diff)
			minerDifficulty = 8
		}
		shareDiff = big.NewInt(int64(minerDifficulty))
	} else {
		shareDiff = util.TargetStrToDiff(t.Target)
	}

	share := Block{
		header:      params[0],
		difficulty:  new(big.Int).Div(util.PowLimit, shareDiff),
	}

	block := Block{
		header:      params[0],
		difficulty:  t.Difficulty,
	}

	if share.Verify() {
		m.heartbeat()
		m.storeShare(shareDiff.Int64())
		atomic.AddUint64(&m.validShares, 1)
		// Log round share for solo mode only
		if !rpc.Pool {
			atomic.AddInt64(&s.roundShares, shareDiff.Int64())
		}
		log.Printf("Valid share from %s@%s at difficulty %v", m.Id, m.IP, shareDiff)
	} else {
		atomic.AddUint64(&m.invalidShares, 1)
		log.Printf("Invalid share from %s@%s", m.Id, m.IP)
		return false
	}

	if rpc.Pool || block.Verify() {
		_, err := rpc.SubmitBlock(paramsOrig)
		now := util.MakeTimestamp()
		if err != nil {
			atomic.AddUint64(&m.rejects, 1)
			atomic.AddUint64(&rpc.Rejects, 1)
			log.Printf("Upstream submission failure on height %v: %v", t.Height, err)
		} else {
			if !rpc.Pool {
				// Solo block found, must refresh job
				s.fetchBlockTemplate()

				// Log this round variance
				roundShares := atomic.SwapInt64(&s.roundShares, 0)
				variance := float64(roundShares) / float64(t.Difficulty.Int64())
				s.blocksMu.Lock()
				s.blockStats[now] = variance
				s.blocksMu.Unlock()
			}
			atomic.AddUint64(&m.accepts, 1)
			atomic.AddUint64(&rpc.Accepts, 1)
			atomic.StoreInt64(&rpc.LastSubmissionAt, now)
			log.Printf("Upstream share found by miner %v@%v at height %d", m.Id, m.IP, t.Height)
		}
	}
	return true
}
