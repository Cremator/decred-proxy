package proxy

import (
	"log"
	"strconv"

	"../util"
)

func (s *ProxyServer) handleGetWorkRPC(cs *Session, diff, id string) (reply *JobReplyData, errorReply *ErrorReply) {
	t := s.currentBlockTemplate()
	if len(t.Header) == 0 {
		return nil, &ErrorReply{Code: -1, Message: "Work not ready"}
	}
	targetHex := t.Target

	if !s.rpc().Pool && diff != ""{
		minerDifficulty, err := strconv.ParseFloat(diff, 64)
		if err != nil {
			log.Printf("Invalid difficulty %v from %v@%v ", diff, id, cs.ip)
			minerDifficulty = 8
		}
		targetHex = util.MakeTargetHex(int64(minerDifficulty))
	}
	reply = &JobReplyData{Data:t.Header, Target:targetHex}
	return
}

func (s *ProxyServer) handleSubmitRPC(cs *Session, diff string, id string, params []string) (reply bool, errorReply *ErrorReply) {
	miner, ok := s.miners.Get(id)
	if !ok {
		miner = NewMiner(id, cs.ip)
		s.registerMiner(miner)
	}

	t := s.currentBlockTemplate()
	reply = miner.processShare(s, t, diff, params)
	return
}

func (s *ProxyServer) handleUnknownRPC(cs *Session, req *JSONRpcReq) *ErrorReply {
	log.Printf("Unknown RPC method: %v", req)
	return &ErrorReply{Code: -1, Message: "Invalid method"}
}
