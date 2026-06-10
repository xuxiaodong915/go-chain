// ============================================================================
// 第九课：完整区块链系统 (Complete Blockchain System)
// ============================================================================
// 目标：将前面所有课程的知识整合成一个完整的区块链系统
// 前置知识：前八课的所有内容
// 学习时长：60分钟（或更多）
//
// 本课整合的功能：
// ================
// 1. ✅ 区块结构（第一课）
// 2. ✅ 哈希计算（第二课）
// 3. ✅ 区块链结构（第三课）
// 4. ✅ 挖矿与工作量证明（第四课）
// 5. ✅ 交易系统（第五课）
// 6. ✅ 数字签名与钱包（第六课）
// 7. ✅ 智能合约（第七课）
// 8. ✅ P2P 网络（第八课）
//
// 本代码可以直接运行，演示一个功能完整的区块链系统！
// ============================================================================

//go:build lesson09

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// 第一课：区块结构
// ============================================================================

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
	Nonce        int64         `json:"nonce"`
}

func (b *Block) calculateHash() string {
	txData := ""
	for _, tx := range b.Transactions {
		txData += tx.From + tx.To + fmt.Sprintf("%.2f", tx.Amount) + tx.Signature + strconv.FormatInt(tx.Timestamp, 10)
	}

	record := fmt.Sprintf("%d%d%s%s%d",
		b.Index, b.Timestamp, txData, b.PreviousHash, b.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func (b *Block) mineBlock(difficulty int) {
	target := strings.Repeat("0", difficulty)
	for {
		hash := b.calculateHash()
		if hash[:difficulty] == target {
			b.Hash = hash
			return
		}
		b.Nonce++
	}
}

// ============================================================================
// 第二课：区块链结构
// ============================================================================

type Blockchain struct {
	Blocks         []*Block
	PendingTxns    []Transaction
	Balances       map[string]float64
	ContractRegs   *ContractRegistry
	mu             sync.RWMutex
	miningReward   float64
	difficulty     int
}

func NewBlockchain(miningReward, difficulty int) *Blockchain {
	genesis := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{
			{From: "genesis", To: "genesis", Amount: 0, Signature: "", Timestamp: time.Now().Unix()},
		},
		PreviousHash: "0",
	}
	genesis.Hash = genesis.calculateHash()

	return &Blockchain{
		Blocks:       []*Block{genesis},
		PendingTxns:  []Transaction{},
		Balances:     make(map[string]float64),
		ContractRegs: NewContractRegistry(),
		miningReward: float64(miningReward),
		difficulty:   difficulty,
	}
}

func (bc *Blockchain) CreateTransaction(from, to string, amount float64, signature string) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if from != "genesis" {
		if bc.Balances[from] < amount {
			return false
		}
	}

	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Signature: signature,
		Timestamp: time.Now().Unix(),
	}
	bc.PendingTxns = append(bc.PendingTxns, tx)
	return true
}

func (bc *Blockchain) MinePendingTransactions(minerAddress string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	allTxns := append([]Transaction{}, bc.PendingTxns...)
	rewardTx := Transaction{
		From:      "genesis",
		To:        minerAddress,
		Amount:    bc.miningReward,
		Signature: "",
		Timestamp: time.Now().Unix(),
	}
	allTxns = append(allTxns, rewardTx)

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: allTxns,
		PreviousHash: prevBlock.Hash,
		Nonce:        0,
	}

	fmt.Printf("⛏️  挖矿区块 #%d...\n", newBlock.Index)
	newBlock.mineBlock(bc.difficulty)

	for _, tx := range allTxns {
		if tx.From != "genesis" {
			bc.Balances[tx.From] -= tx.Amount
		}
		bc.Balances[tx.To] += tx.Amount
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingTxns = []Transaction{}
	return newBlock
}

func (bc *Blockchain) Validate() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		if current.Hash != current.calculateHash() {
			return false
		}
		if current.PreviousHash != previous.Hash {
			return false
		}
	}
	return true
}

func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Balances[address]
}

// ============================================================================
// 交易结构
// ============================================================================

type Transaction struct {
	From      string  `json:"from"`
	To        string  `json:"to"`
	Amount    float64 `json:"amount"`
	Signature string  `json:"signature"`
	Timestamp int64   `json:"timestamp"`
}

// ============================================================================
// 第六课：钱包
// ============================================================================

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

func NewWallet() *Wallet {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privateKey.PublicKey.X, privateKey.PublicKey.Y)
	h := sha256.New()
	h.Write(pubKeyBytes)
	address := hex.EncodeToString(h.Sum(nil))

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Address:    address,
	}
}

func (w *Wallet) Sign(data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(append(r.Bytes(), s.Bytes()...)), nil
}

func (w *Wallet) SignTransaction(tx *Transaction) error {
	txData := fmt.Sprintf("%s%s%.2f%d", tx.From, tx.To, tx.Amount, tx.Timestamp)
	sig, err := w.Sign([]byte(txData))
	if err != nil {
		return err
	}
	tx.Signature = sig
	return nil
}

// ============================================================================
// 第七课：智能合约
// ============================================================================

type Contract interface {
	Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{}
	GetName() string
}

type ContractContext struct {
	Storage map[string]string
	Caller  string
}

type TokenContract struct {
	name  string
	owner string
}

func NewTokenContract(name, owner string) *TokenContract {
	return &TokenContract{name: name, owner: owner}
}

func (t *TokenContract) GetName() string {
	return t.name
}

func (t *TokenContract) Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{} {
	switch method {
	case "balanceOf":
		address := params["address"].(string)
		if bal, ok := ctx.Storage["balance_"+address]; ok {
			return map[string]interface{}{"balance": bal}
		}
		return map[string]interface{}{"balance": "0"}
	case "transfer":
		from := params["from"].(string)
		to := params["to"].(string)
		amount := params["amount"].(string)
		fromBal, _ := strconv.ParseFloat(ctx.Storage["balance_"+from], 64)
		amt, _ := strconv.ParseFloat(amount, 64)
		if fromBal >= amt {
			toBal, _ := strconv.ParseFloat(ctx.Storage["balance_"+to], 64)
			ctx.Storage["balance_"+from] = strconv.FormatFloat(fromBal-amt, 'f', -1, 64)
			ctx.Storage["balance_"+to] = strconv.FormatFloat(toBal+amt, 'f', -1, 64)
			return map[string]interface{}{"success": true}
		}
		return map[string]interface{}{"success": false, "error": "insufficient"}
	case "mint":
		if ctx.Caller != t.owner {
			return map[string]interface{}{"error": "not owner"}
		}
		addr := params["address"].(string)
		amount := params["amount"].(string)
		ctx.Storage["balance_"+addr] = amount
		return map[string]interface{}{"success": true}
	}
	return map[string]interface{}{"error": "unknown method"}
}

type ContractRegistry struct {
	contracts map[string]Contract
	storages  map[string]map[string]string
}

func NewContractRegistry() *ContractRegistry {
	return &ContractRegistry{
		contracts: make(map[string]Contract),
		storages:  make(map[string]map[string]string),
	}
}

func (cr *ContractRegistry) Deploy(name string, contract Contract) {
	cr.contracts[name] = contract
	cr.storages[name] = make(map[string]string)
}

func (cr *ContractRegistry) Execute(name string, caller string, method string, params map[string]interface{}) interface{} {
	contract, ok := cr.contracts[name]
	if !ok {
		return map[string]interface{}{"error": "not found"}
	}
	ctx := &ContractContext{
		Storage: cr.storages[name],
		Caller:  caller,
	}
	return contract.Execute(ctx, method, params)
}

// ============================================================================
// 第八课：P2P 节点
// ============================================================================

type P2PNode struct {
	ID         string
	Address    string
	Peers      []string
	Blockchain *Blockchain
	mu         sync.RWMutex
}

func NewP2PNode(id, address string, bc *Blockchain) *P2PNode {
	return &P2PNode{
		ID:         id,
		Address:    address,
		Peers:      []string{},
		Blockchain: bc,
	}
}

func (n *P2PNode) AddPeer(addr string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Peers = append(n.Peers, addr)
}

func (n *P2PNode) Broadcast(msgType string, payload map[string]interface{}) {
	n.mu.RLock()
	peers := make([]string, len(n.Peers))
	copy(peers, n.Peers)
	n.mu.RUnlock()

	msg := map[string]interface{}{
		"type":      msgType,
		"from":      n.ID,
		"timestamp": time.Now().Unix(),
		"payload":   payload,
	}
	data, _ := json.Marshal(msg)

	for _, peer := range peers {
		go func(addr string) {
			http.Post("http://"+addr+"/receive", "application/json", strings.NewReader(string(data)))
		}(peer)
	}
}

// ============================================================================
// 主函数：完整演示
// ============================================================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🚀 完整区块链系统演示")
	fmt.Println(strings.Repeat("=", 60))

	// 创建区块链（挖矿奖励 10，难度 2）
	blockchain := NewBlockchain(10, 2)

	// 创建钱包
	fmt.Println("\n💼 创建钱包...")
	alice := NewWallet()
	bob := NewWallet()
	charlie := NewWallet()

	fmt.Printf("  Alice: %s\n", alice.Address[:16]+"...")
	fmt.Printf("  Bob: %s\n", bob.Address[:16]+"...")
	fmt.Printf("  Charlie: %s\n", charlie.Address[:16]+"...")

	// 初始化余额
	blockchain.Balances[alice.Address] = 100
	blockchain.Balances[bob.Address] = 50

	fmt.Println("\n💰 初始余额:")
	fmt.Printf("  Alice: %.2f\n", blockchain.GetBalance(alice.Address))
	fmt.Printf("  Bob: %.2f\n", blockchain.GetBalance(bob.Address))
	fmt.Printf("  Charlie: %.2f\n", blockchain.GetBalance(charlie.Address))

	// 创建交易
	fmt.Println("\n💸 创建交易...")
	tx1 := Transaction{From: alice.Address, To: bob.Address, Amount: 30, Timestamp: time.Now().Unix()}
	alice.SignTransaction(&tx1)
	blockchain.CreateTransaction(alice.Address, bob.Address, 30, tx1.Signature)

	tx2 := Transaction{From: bob.Address, To: charlie.Address, Amount: 20, Timestamp: time.Now().Unix()}
	bob.SignTransaction(&tx2)
	blockchain.CreateTransaction(bob.Address, charlie.Address, 20, tx2.Signature)

	fmt.Println("✅ 2 笔交易已加入待确认池")

	// 挖矿确认
	fmt.Println("\n⛏️  挖矿确认交易...")
	blockchain.MinePendingTransactions(alice.Address)

	fmt.Println("\n💰 挖矿后余额:")
	fmt.Printf("  Alice: %.2f (+10 矿工奖励)\n", blockchain.GetBalance(alice.Address))
	fmt.Printf("  Bob: %.2f\n", blockchain.GetBalance(bob.Address))
	fmt.Printf("  Charlie: %.2f\n", blockchain.GetBalance(charlie.Address))

	// 智能合约
	fmt.Println("\n📜 智能合约演示...")
	token := NewTokenContract("MyToken", alice.Address)
	blockchain.ContractRegs.Deploy("MyToken", token)

	result := blockchain.ContractRegs.Execute("MyToken", alice.Address, "mint",
		map[string]interface{}{"address": alice.Address, "amount": "1000"})
	fmt.Println("  Mint 1000:", result)

	result = blockchain.ContractRegs.Execute("MyToken", alice.Address, "balanceOf",
		map[string]interface{}{"address": alice.Address})
	fmt.Println("  Alice Balance:", result)

	// 验证区块链
	fmt.Println("\n🔍 验证区块链...")
	if blockchain.Validate() {
		fmt.Println("✅ 区块链验证通过！")
	} else {
		fmt.Println("❌ 区块链验证失败！")
	}

	// 打印区块链
	fmt.Println("\n📊 区块链信息:")
	for _, block := range blockchain.Blocks {
		fmt.Printf("  区块 #%d: %d 笔交易, 哈希=%s\n",
			block.Index, len(block.Transactions), block.Hash[:16]+"...")
	}

	// P2P 节点
	fmt.Println("\n🌐 P2P 网络演示...")
	node1 := NewP2PNode("Node1", "localhost:8080", blockchain)
	node2 := NewP2PNode("Node2", "localhost:8081", blockchain)

	node1.AddPeer("localhost:8081")
	node2.AddPeer("localhost:8080")

	fmt.Printf("  Node1 有 %d 个对等节点\n", node1.GetPeerCount())

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("✅ 完整演示结束！")
	fmt.Println(strings.Repeat("=", 60))

	fmt.Println("\n🎉 你已经学习了区块链的核心概念：")
	fmt.Println("  ✅ 区块和哈希")
	fmt.Println("  ✅ 挖矿和共识")
	fmt.Println("  ✅ 交易和钱包")
	fmt.Println("  ✅ 智能合约")
	fmt.Println("  ✅ P2P 网络")
}

// GetPeerCount 辅助方法
func (n *P2PNode) GetPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.Peers)
}