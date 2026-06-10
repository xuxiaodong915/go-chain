// ============================================================================
// 第五课：交易系统 (Transaction System)
// ============================================================================
// 目标：理解区块链中的交易结构和交易处理流程
// 前置知识：第一课 - 区块结构，第二课 - 哈希计算，第四课 - 挖矿
// 学习时长：30分钟
//
// 什么是交易？
// ============
// 在区块链中，交易是价值转移的基本单位。
// 比如：Alice 转账 50 BTC 给 Bob，这就是一笔交易。
//
// 交易的基本结构
// ===============
// 1. From: 发送方地址
// 2. To: 接收方地址
// 3. Amount: 转账金额
// 4. Timestamp: 交易时间戳
// 5. Signature: 数字签名（下一课会讲）
//
// 交易的生命周期
// ===============
// 1. 创建：用户发起交易
// 2. 广播：交易发送到网络中的所有节点
// 3. 验证：节点验证交易是否有效（余额、签名等）
// 4. 打包：矿工将有效交易打包进区块
// 5. 确认：区块被添加到区块链，交易被确认
//
// 待确认交易池（Mempool）
// =======================
// 在打包进区块之前，交易会被放在"待确认交易池"中。
// 每个节点都有自己的 Mempool。
// ============================================================================

//go:build lesson05

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"   // 用于并发安全
	"time"
)

// ============================================================================
// 交易结构
// ============================================================================

// Transaction 交易结构
type Transaction struct {
	From      string  // 发送方地址（"genesis" 表示系统奖励）
	To        string  // 接收方地址
	Amount    float64 // 转账金额
	Timestamp int64   // 交易时间戳
}

// String 实现 Stringer 接口，方便打印
func (t *Transaction) String() string {
	return fmt.Sprintf("%s -> %.2f -> %s", t.From, t.Amount, t.To)
}

// ============================================================================
// 区块结构（支持交易）
// ============================================================================

// Block 区块结构 - 使用交易切片替代简单的字符串
type Block struct {
	Index        int64         // 区块编号
	Timestamp    int64         // 时间戳
	Transactions []Transaction // 交易列表（不再用简单的字符串）
	PreviousHash string        // 前一个区块的哈希
	Hash         string        // 当前区块的哈希
	Nonce        int64         // 挖矿用的随机数
}

// calculateHash 计算区块哈希（包含所有交易）
func (b *Block) calculateHash() string {
	// 将所有交易序列化为字符串
	txString := ""
	for _, tx := range b.Transactions {
		txString += fmt.Sprintf("%s%s%.2f%d", tx.From, tx.To, tx.Amount, tx.Timestamp)
	}

	record := fmt.Sprintf("%d%d%s%s%d",
		b.Index,
		b.Timestamp,
		txString,
		b.PreviousHash,
		b.Nonce)

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
// 区块链结构（支持交易）
// ============================================================================

type Blockchain struct {
	Blocks         []*Block         // 所有区块
	PendingTxns    []Transaction    // 待确认交易池（Mempool）
	Balances       map[string]float64 // 账户余额
	mu             sync.RWMutex     // 读写锁（并发安全）
	miningReward   float64          // 挖矿奖励
}

// NewBlockchain 创建新区块链
func NewBlockchain(miningReward float64) *Blockchain {
	// 创建创世区块（包含一笔系统交易）
	genesis := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{
			{From: "genesis", To: "genesis", Amount: 0, Timestamp: time.Now().Unix()},
		},
		PreviousHash: "0",
		Nonce:        0,
	}
	genesis.Hash = genesis.calculateHash()

	// 初始化余额（给一些测试账户初始余额）
	balances := map[string]float64{
		"Alice":   1000,
		"Bob":     500,
		"Charlie": 200,
		"Dave":    100,
	}

	return &Blockchain{
		Blocks:       []*Block{genesis},
		PendingTxns:  []Transaction{},
		Balances:     balances,
		miningReward: miningReward,
	}
}

// GetBalance 获取账户余额
func (bc *Blockchain) GetBalance(address string) float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Balances[address]
}

// ============================================================================
// 核心方法：CreateTransaction 创建交易
// ============================================================================

// CreateTransaction 创建新交易并添加到待确认池
//
// 参数：
//   - from: 发送方地址
//   - to: 接收方地址
//   - amount: 转账金额
//
// 返回：
//   - bool: 是否成功创建交易
//   - string: 错误信息（如果失败）
//
// 交易验证规则：
// ==============
// 1. 发送方余额必须 >= 转账金额
// 2. 金额必须为正数
// 3. 发送方和接收方不能相同
func (bc *Blockchain) CreateTransaction(from, to string, amount float64) (bool, string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// 验证1：金额必须为正数
	if amount <= 0 {
		return false, "金额必须大于0"
	}

	// 验证2：不能给自己转账
	if from == to {
		return false, "不能给自己转账"
	}

	// 验证3：检查余额（genesis 交易例外）
	if from != "genesis" {
		balance, exists := bc.Balances[from]
		if !exists {
			return false, fmt.Sprintf("发送方账户 %s 不存在", from)
		}
		if balance < amount {
			return false, fmt.Sprintf("余额不足！当前余额: %.2f, 需要: %.2f", balance, amount)
		}
	}

	// 创建交易
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}

	// 添加到待确认池
	bc.PendingTxns = append(bc.PendingTxns, tx)

	return true, "交易创建成功，已加入待确认池"
}

// GetPendingTxnsCount 获取待确认交易数量
func (bc *Blockchain) GetPendingTxnsCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.PendingTxns)
}

// PrintPendingTxns 打印待确认交易
func (bc *Blockchain) PrintPendingTxns() {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.PendingTxns) == 0 {
		fmt.Println("  📭 待确认池为空")
		return
	}

	fmt.Println("  📋 待确认交易:")
	for i, tx := range bc.PendingTxns {
		fmt.Printf("    [%d] %s\n", i+1, tx.String())
	}
}

// ============================================================================
// 核心方法：MinePendingTransactions 挖矿确认交易
// ============================================================================

// MinePendingTransactions 挖矿确认所有待确认交易
//
// 参数：
//   - minerAddress: 矿工地址（挖矿奖励会发到这个地址）
//
// 返回：
//   - *Block: 新挖出的区块
//
// 挖矿步骤：
// 1. 将所有待确认交易打包
// 2. 添加一笔挖矿奖励交易（From=genesis, To=矿工地址）
// 3. 创建新区块并挖矿
// 4. 更新账户余额
// 5. 清空待确认池
func (bc *Blockchain) MinePendingTransactions(minerAddress string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// 第1步：准备所有交易
	// 包括：待确认交易 + 挖矿奖励交易
	allTxns := make([]Transaction, len(bc.PendingTxns))
	copy(allTxns, bc.PendingTxns)

	// 第2步：添加挖矿奖励
	rewardTx := Transaction{
		From:      "genesis",
		To:        minerAddress,
		Amount:    bc.miningReward,
		Timestamp: time.Now().Unix(),
	}
	allTxns = append(allTxns, rewardTx)

	// 第3步：创建新区块
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index:        prevBlock.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: allTxns,
		PreviousHash: prevBlock.Hash,
		Nonce:        0,
	}

	// 第4步：挖矿
	fmt.Printf("\n⛏️  矿工 %s 开始挖矿，确认 %d 笔交易...\n",
		minerAddress, len(bc.PendingTxns))
	newBlock.mineBlock(2)
	fmt.Printf("✅ 区块 #%d 挖矿成功！\n", newBlock.Index)

	// 第5步：更新余额
	for _, tx := range allTxns {
		if tx.From != "genesis" {
			// 扣除发送方余额
			bc.Balances[tx.From] -= tx.Amount
		}
		// 增加接收方余额
		bc.Balances[tx.To] += tx.Amount
	}

	// 第6步：添加区块并清空待确认池
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingTxns = []Transaction{}

	return newBlock
}

// ============================================================================
// 辅助方法
// ============================================================================

func (bc *Blockchain) PrintBalances() {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	fmt.Println("\n💰 当前账户余额:")
	fmt.Println("   " + strings.Repeat("-", 40))
	for addr, balance := range bc.Balances {
		fmt.Printf("   %-12s: %.2f BTC\n", addr, balance)
	}
	fmt.Println("   " + strings.Repeat("-", 40))
}

func (bc *Blockchain) PrintBlockchain() {
	fmt.Println("\n📊 区块链信息:")
	for _, block := range bc.Blocks {
		fmt.Printf("\n区块 #%d\n", block.Index)
		fmt.Printf("  时间: %s\n", time.Unix(block.Timestamp, 0).Format("15:04:05"))
		fmt.Printf("  交易数: %d\n", len(block.Transactions))
		fmt.Printf("  哈希: %s\n", block.Hash[:16]+"...")
		if len(block.Transactions) > 0 && block.Transactions[0].From != "genesis" {
			fmt.Printf("  首笔交易: %s\n", block.Transactions[0].String())
		}
	}
}

// ============================================================================
// 主函数：演示交易系统
// ============================================================================

func main() {
	fmt.Println("📚 第五课：交易系统")
	fmt.Println("=" + strings.Repeat("=", 49))

	// 创建区块链，挖矿奖励设为 10 BTC
	blockchain := NewBlockchain(10)

	// -------------------------------------------------------------------------
	// 演示1：查看初始余额
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：初始账户余额")
	fmt.Println("----------------------------------------")
	blockchain.PrintBalances()

	// -------------------------------------------------------------------------
	// 演示2：创建交易
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：创建交易")
	fmt.Println("----------------------------------------")

	// Alice 转账给 Bob
	ok, msg := blockchain.CreateTransaction("Alice", "Bob", 50)
	if ok {
		fmt.Printf("✅ %s\n", msg)
	} else {
		fmt.Printf("❌ %s\n", msg)
	}

	// Bob 转账给 Charlie
	ok, msg = blockchain.CreateTransaction("Bob", "Charlie", 30)
	if ok {
		fmt.Printf("✅ %s\n", msg)
	} else {
		fmt.Printf("❌ %s\n", msg)
	}

	// 查看待确认交易
	fmt.Printf("\n待确认交易数: %d\n", blockchain.GetPendingTxnsCount())
	blockchain.PrintPendingTxns()

	// -------------------------------------------------------------------------
	// 演示3：余额不足的交易
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：余额不足的交易")
	fmt.Println("----------------------------------------")

	ok, msg = blockchain.CreateTransaction("Dave", "Alice", 200)
	if ok {
		fmt.Printf("✅ %s\n", msg)
	} else {
		fmt.Printf("❌ %s\n", msg)
	}

	// 查看当前余额
	blockchain.PrintBalances()

	// -------------------------------------------------------------------------
	// 演示4：挖矿确认交易
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：挖矿确认交易")
	fmt.Println("----------------------------------------")

	// Alice 作为矿工挖矿
	blockchain.MinePendingTransactions("Alice")

	// 查看挖矿后的余额
	fmt.Println("\n💰 挖矿后的余额 (Alice 获得 10 BTC 奖励):")
	blockchain.PrintBalances()

	// -------------------------------------------------------------------------
	// 演示5：更多交易
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示5：更多交易")
	fmt.Println("----------------------------------------")

	blockchain.CreateTransaction("Charlie", "Dave", 15)
	blockchain.CreateTransaction("Alice", "Bob", 25)

	fmt.Println("\n待确认交易:")
	blockchain.PrintPendingTxns()

	// Bob 挖矿
	blockchain.MinePendingTransactions("Bob")

	fmt.Println("\n💰 最终余额:")
	blockchain.PrintBalances()

	// -------------------------------------------------------------------------
	// 演示6：查看区块链
	// -------------------------------------------------------------------------
	blockchain.PrintBlockchain()

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. 交易是价值转移的基本单位，包含发送方、接收方、金额等信息
	// 2. 交易创建后会进入待确认池（Mempool）
	// 3. 挖矿会将待确认交易打包进区块并确认
	// 4. 矿工会获得挖矿奖励
	// 5. 系统需要验证交易的余额是否充足
	//
	// ❓ 思考题：
	// 1. 如果多个矿工同时挖矿，谁获得的奖励？
	// 2. 如果待确认池中有1000笔交易，全部打包可行吗？
	// 3. 为什么要设置交易费用？
	//
	// 🔜 下一课：数字签名 - 学习如何验证交易的真实性
	// ============================================================================
	fmt.Println("\n✅ 第五课完成！下一课将学习数字签名。")
}