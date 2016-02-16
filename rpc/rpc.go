package rpc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type RPCClient struct {
	sync.RWMutex
	Url              *url.URL
	Name             string
	Username         string
	Password         string
	Pool             bool
	sick             bool
	sickRate         int
	successRate      int
	Accepts          uint64
	Rejects          uint64
	LastSubmissionAt int64
	client           *http.Client
	FailsCount       uint64
	Req              uint64
}

const (
	dumpprotocol = true
)

type GetWorkReply struct {
	Data     string `json:"data"`
	Target   string `json:"target"`
}

type GetBlockReply struct {
	Number     string `json:"number"`
	Difficulty string `json:"difficulty"`
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func NewRPCClient(name, rawUrl, username string, password string, timeout string, pool bool) (*RPCClient, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	rpcClient := &RPCClient{Name: name, Url: url, Pool: pool, Username:username, Password:password}
	timeoutIntv, _ := time.ParseDuration(timeout)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
	rpcClient.client = &http.Client{
		Transport: tr,
		Timeout: timeoutIntv,
	}
	return rpcClient, nil
}

func (r *RPCClient) GetWork() (GetWorkReply, error) {
	params := []string{}

	rpcResp, err := r.doPost(r.Url.String(), "getwork", params)
	var reply GetWorkReply
	if err != nil {
		return reply, err
	}
	if rpcResp.Error != nil {
		return reply, errors.New(rpcResp.Error["message"].(string))
	}

	err = json.Unmarshal(*rpcResp.Result, &reply)
	//fmt.Printf("R: %T %+v\n", reply, reply)
	// Handle empty result, daemon is catching up (geth bug!!!)
	if err != nil {
		return reply, errors.New("Daemon is not ready")
	}
	return reply, err
}

func (r *RPCClient) SubmitBlock(params []string) (bool, error) {
	rpcResp, err := r.doPost(r.Url.String(), "getwork", params)
	var result bool
	if err != nil {
		return false, err
	}
	if rpcResp.Error != nil {
		return false, errors.New(rpcResp.Error["message"].(string))
	}
	err = json.Unmarshal(*rpcResp.Result, &result)
	if !result {
		return false, errors.New("Block not accepted, result=false")
	}
	return result, nil
}

func (r *RPCClient) doPost(url, method string, params interface{}) (JSONRpcResp, error) {
	r.Req++
	reqcount := r.Req
	jsonReq := map[string]interface{}{"id": 0, "method": method, "params": params}
	data, _ := json.Marshal(jsonReq)
	if dumpprotocol {
		fmt.Printf("Send(%d): %s\n", reqcount, data)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.SetBasicAuth(r.Username, r.Password)
	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if dumpprotocol {
		dump, err2 := httputil.DumpRequestOut(req, true)
		fmt.Printf("REQ: %+v %+v\n", string(dump), err2)
	}
	resp, err := r.client.Do(req)
	if dumpprotocol {
		dump, err2 := httputil.DumpResponse(resp, true)
		fmt.Printf("RES: %+v %+v\n", string(dump), err2)
	}
	var rpcResp JSONRpcResp

	if err != nil {
		r.markSick()
		return rpcResp, err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if dumpprotocol {
		fmt.Printf("Recv(%d): %s\n", reqcount, body)
	}
	err = json.Unmarshal(body, &rpcResp)

	if rpcResp.Error != nil {
		r.markSick()
	}
	return rpcResp, err
}

func (r *RPCClient) Check() (bool, error) {
	_, err := r.GetWork()
	if err != nil {
		return false, err
	}
	r.markAlive()
	return !r.Sick(), nil
}

func (r *RPCClient) Sick() bool {
	r.RLock()
	defer r.RUnlock()
	return r.sick
}

func (r *RPCClient) markSick() {
	r.Lock()
	if !r.sick {
		atomic.AddUint64(&r.FailsCount, 1)
	}
	r.sickRate++
	r.successRate = 0
	if r.sickRate >= 5 {
		r.sick = true
	}
	r.Unlock()
}

func (r *RPCClient) markAlive() {
	r.Lock()
	r.successRate++
	if r.successRate >= 5 {
		r.sick = false
		r.sickRate = 0
		r.successRate = 0
	}
	r.Unlock()
}
