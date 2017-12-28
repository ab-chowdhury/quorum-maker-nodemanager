package client

import (
	"fmt"
	"github.com/ybbus/jsonrpc"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"strconv"
	"log"
	"io/ioutil"
	"strings"
	"github.com/magiconair/properties"
	"regexp"
)

type NodeInfo struct {
    ConnectionInfo  ConnectionInfo  `json:"connectionInfo,omitempty"`
    RaftRole        string          `json:"raftRole,omitempty"`
    RaftID          int             `json:"raftID,omitempty"`
    BlockNumber     int64           `json:"blockNumber,omitempty"`
    PendingTxCount  int             `json:"pendingTxCount"`
    Genesis         string          `json:"genesis,omitempty"`
    AdminInfo       AdminInfo       `json:"adminInfo,omitempty"`
}

type ConnectionInfo struct {
    IP      string  `json:"ip,omitempty"`
    Port    int     `json:"port,omitempty"`
    Enode   string  `json:"enode,omitempty"`
}

type AdminInfo struct {
    ID          string      `json:"id,omitempty"`
    Name        string      `json:"name,omitempty"`
    Enode       string      `json:"enode,omitempty"`
    IP          string      `json:"ip,omitempty"`
    Ports       Ports       `json:"ports,omitempty"`
    ListenAddr  string      `json:"listenAddr,omitempty"`
    Protocols   Protocols   `json:"protocols,omitempty"`
}

type Ports struct {
	Discovery int `json:"discovery,omitempty"`
	Listener  int `json:"listener,omitempty"`
}

type AdminPeers struct {
	ID      	string   	`json:"id,omitempty"`
	Name    	string   	`json:"name,omitempty"`
	Caps    	[]string 	`json:"caps,omitempty"`
	Network 	Network 	`json:"network,omitempty"`
	Protocols 	Protocols 	`json:"protocols,omitempty"`
}

type Protocols struct {
	Eth Eth `json:"eth,omitempty"`
}

type Eth struct {
	Network    int	  `json:"network,omitempty"`
 	Version    int    `json:"version,omitempty"`
	Difficulty int    `json:"difficulty,omitempty"`
	Genesis    string `json:"genesis,omitempty"`
	Head       string `json:"head,omitempty"`
}

type Network struct {
	LocalAddress  string `json:"localAddress,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
}

type BlockDetailsResponse struct {
	Number           string                       `json:"number"`
	Hash             string                       `json:"hash"`
	ParentHash       string                       `json:"parentHash"`
	Nonce            string                       `json:"nonce"`
	Sha3Uncles       string                       `json:"sha3Uncles"`
	LogsBloom        string                       `json:"logsBloom"`
	TransactionsRoot string                       `json:"transactionsRoot"`
	StateRoot        string                       `json:"stateRoot"`
	Miner            string                       `json:"miner"`
	Difficulty       string                       `json:"difficulty"`
	TotalDifficulty  string                       `json:"totalDifficulty"`
	ExtraData        string                       `json:"extraData"`
	Size             string                       `json:"size"`
	GasLimit         string                       `json:"gasLimit"`
	GasUsed          string                       `json:"gasUsed"`
	Timestamp        string                       `json:"timestamp"`
	Transactions     []TransactionDetailsResponse `json:"transactions"`
	Uncles           []string                     `json:"uncles"`
}

type TransactionDetailsResponse struct {
	BlockHash        string `json:"blockHash,omitempty"`
	BlockNumber      string `json:"blockNumber,omitempty"`
	From             string `json:"from,omitempty"`
	Gas              string `json:"gas,omitempty"`
	GasPrice         string `json:"gasPrice,omitempty"`
	Hash             string `json:"hash,omitempty"`
	Input            string `json:"input,omitempty"`
	Nonce            string `json:"nonce,omitempty"`
	To               string `json:"to,omitempty"`
	TransactionIndex string `json:"transactionIndex,omitempty"`
	Value            string `json:"value,omitempty"`
	V                string `json:"v,omitempty"`
	R                string `json:"r,omitempty"`
	S                string `json:"s,omitempty"`
}

type EthClient struct {
	Url string
}

func (ec *EthClient) GetTransactionInfo(txno string) (TransactionDetailsResponse) {
	rpcClient := jsonrpc.NewRPCClient(ec.Url)
	response, err := rpcClient.Call("eth_getTransactionByHash", txno)

	if err != nil {
		fmt.Println(err)
	}

	txresponse := TransactionDetailsResponse{}

	err = response.GetObject(&txresponse)
	if err != nil {
		fmt.Println(err)
	}
	return txresponse
}

func (ec *EthClient) GetTransactionInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if params["id"] == "pending" {
		response := ec.GetPendingTransactions()
		json.NewEncoder(w).Encode(response)
	} else {
		response := ec.GetTransactionInfo(params["id"])
		json.NewEncoder(w).Encode(response)
	}
}

func (ec *EthClient) GetBlockInfo(blockno int64) (BlockDetailsResponse) {
	rpcClient := jsonrpc.NewRPCClient(ec.Url)
	blocknohex  := strconv.FormatInt(blockno, 16)
	bnohex := fmt.Sprint("0x", blocknohex)

	response, err := rpcClient.Call("eth_getBlockByNumber", bnohex, true)
	if err != nil {
		fmt.Println(err)
	}

	blockresponse := BlockDetailsResponse{}
	err = response.GetObject(&blockresponse)
	if err != nil {
		fmt.Println(err)
	}
	return blockresponse
}

func (ec *EthClient) GetBlockInfoHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	block, err := strconv.ParseInt(params["id"], 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	response := ec.GetBlockInfo(block)
	fmt.Print(response)
	json.NewEncoder(w).Encode(response)
}

func (ec *EthClient) GetPendingTransactions() ([]TransactionDetailsResponse) {
	rpcClient := jsonrpc.NewRPCClient(ec.Url)
	response, err := rpcClient.Call("eth_pendingTransactions")
	if err != nil {
		fmt.Println(err)
	}
	pendingtxresponse := []TransactionDetailsResponse{}
	err = response.GetObject(&pendingtxresponse)
	if err != nil {
		fmt.Println(err)
	}
	return pendingtxresponse
}

func (ec *EthClient) GetOtherPeer(peerid string) (AdminPeers) {
	rpcClient := jsonrpc.NewRPCClient(ec.Url)
	response, err := rpcClient.Call("admin_peers")
	if err != nil {
		fmt.Println(err)
	}
	otherpeersresponse := []AdminPeers{}
	err = response.GetObject(&otherpeersresponse)
	if err != nil {
		fmt.Println(err)
	}
	for _, item := range otherpeersresponse {
		if item.ID == peerid {
			peerresponse := item
			return peerresponse
		}
	}
	return AdminPeers{}
}

func (ec *EthClient) GetOtherPeerHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	response := ec.GetOtherPeer(params["id"])
	fmt.Print(response)
	json.NewEncoder(w).Encode(response)
}

func (ec *EthClient) GetCurrentNode () (NodeInfo) {
	p := properties.MustLoadFile("/home/setup.conf", properties.UTF8)
	var filename string
	ipaddr := p.MustGetString("CURRENT_IP")
	rpcport := p.MustGetString("RPC_PORT")

	//Alternate regex that can be used is (start_)(\w)*.sh
	r, _ := regexp.Compile("[s][t][a][r][t][_][A-Za-z0-9]*[.][s][h]")
	files, err := ioutil.ReadDir("/home/node")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		match, _ := regexp.MatchString("[s][t][a][r][t][_][A-Za-z0-9]*[.][s][h]", f.Name())
		if(match) {
			filename = r.FindString(f.Name())
		}
	}

	filepath := fmt.Sprint("/home/node/", filename)
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
	}

	lines := strings.Split(string(content), "\n")
	raftidline := lines[4]
	lines = strings.Split(string(raftidline), "=")
	raftid := lines[1]

	rpcClient := jsonrpc.NewRPCClient(ec.Url)
	response, err := rpcClient.Call("admin_nodeInfo")
	if err != nil {
		fmt.Println(err)
	}
	thisadmininfo := AdminInfo{}
	err = response.GetObject(&thisadmininfo)

	enode := thisadmininfo.Enode
	rpcClient = jsonrpc.NewRPCClient(ec.Url)
	response, err = rpcClient.Call("eth_pendingTransactions")
	if err != nil {
		fmt.Println(err)
	}
	pendingtxresponse := []TransactionDetailsResponse{}
	err = response.GetObject(&pendingtxresponse)
	pendingtxcount := len(pendingtxresponse)

	if err != nil {
		fmt.Println(err)
	}

	rpcClient = jsonrpc.NewRPCClient(ec.Url)
	response, err = rpcClient.Call("eth_blockNumber")
	if err != nil {
		fmt.Println(err)
	}
	var blocknumber string;
	err = response.GetObject(&blocknumber)
	if err != nil {
		fmt.Println(err)
	}
	blocknumber = strings.TrimSuffix(blocknumber, "\n")
	blocknumber = strings.TrimPrefix(blocknumber, "0x")
	blocknumberInt, err := strconv.ParseInt(blocknumber, 16, 64)
	if err != nil {
		fmt.Println(err)
	}

	raftid = strings.TrimSuffix(raftid, "\n")

	raftidInt, err := strconv.Atoi(raftid)
	if err != nil {
		log.Fatal(err)
	}

	rpcClient = jsonrpc.NewRPCClient(ec.Url)
	response, err = rpcClient.Call("raft_role")
	if err != nil {
		fmt.Println(err)
	}
	var raftrole string;
	err = response.GetObject(&raftrole)
	if err != nil {
		fmt.Println(err)
	}
	raftrole = strings.TrimSuffix(raftrole, "\n")

	rpcport = strings.TrimSuffix(rpcport, "\n")

	rpcportInt, err := strconv.Atoi(rpcport)
	if err != nil {
		log.Fatal(err)
	}

	ipaddr = strings.TrimSuffix(ipaddr, "\n")
	b, err := ioutil.ReadFile("/home/node/genesis.json")
	if err != nil {
		log.Fatal(err)
	}

	genesis := string(b)
	genesis = strings.Replace(genesis, "\n","",-1)
	conn := ConnectionInfo{ipaddr,rpcportInt,enode}
	responseobj := NodeInfo{conn,raftrole,raftidInt,blocknumberInt,pendingtxcount,genesis,thisadmininfo}
	return responseobj
}

func (ec *EthClient) GetCurrentNodeHandler(w http.ResponseWriter, r *http.Request) {
	response := ec.GetCurrentNode()
	fmt.Print(response)
	json.NewEncoder(w).Encode(response)
}