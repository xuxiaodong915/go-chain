// ============================================================================
// 第三课：区块链结构 (Blockchain Structure)
// ============================================================================
// 目标：理解如何将多个区块组织成一条区块链，并验证其完整性
// 前置知识：第一课 - 区块结构，第二课 - 哈希计算
// 学习时长：25分钟
//
// 什么是区块链？
// ==============
// 区块链是一个由多个区块按顺序连接而成的链表。
// 每个区块都包含前一个区块的哈希值，形成一条不可篡改的链条。
//
// 为什么叫"链"？
// ==============
// 因为一环扣一环：
// 区块0 ← 区块1 ← 区块2 ← 区块3 ← ...
// 每个"←"就是 PreviousHash 指针
//
// 区块链的数据结构
// =================
// type Blockchain struct {
//     Blocks []*Block  // 一个指针切片，存储所有区块
// }
//
// 为什么用 []*Block 而不是 []Block？
// ==================================
// 1. 指针切片更节省内存（只存地址）
// 2. 操作时不需要复制整个结构体
// 3. 符合 Go 的惯用法
// ============================================================================

//go:build lesson03

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Block 区块结构（复用前两课的代码）
type Block struct {
	Index        int64
	Timestamp    int64
	Data         string
	PreviousHash string
	Hash         string
	Nonce        int64
}

func (b *Block) calculateHash() string {
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.Data, b.PreviousHash, b.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// Blockchain 区块链结构
// ============================================================================

type Blockchain struct {
	Blocks []*Block // 存储所有区块的切片
}

// NewBlockchain 创建新的区块链
//
// 返回包含创世区块的区块链
//
// 什么是创世区块？
// ================
// 创世区块（Genesis Block）是区块链的第一个区块，
// 它的 PreviousHash 设为 "0"，表示没有前一个区块。
func NewBlockchain() *Blockchain {
	// 创建创世区块
	genesis := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Data:         "创世区块：区块链的起点",
		PreviousHash: "0", // 创世区块的前哈希固定为 "0"
		Nonce:        0,
	}

	// 计算创世区块的哈希
	genesis.Hash = genesis.calculateHash()

	// 返回包含创世区块的区块链
	return &Blockchain{
		Blocks: []*Block{genesis},
	}
}

// AddBlock 添加新区块到区块链
//
// 参数：
//   - data: 要存储在区块中的数据
//
// 返回：
//   - *Block: 新创建的区块
//
// 添加步骤：
// 1. 获取最后一个区块（当前区块链的末端）
// 2. 创建新区块，其 PreviousHash 指向最后一个区块的 Hash
// 3. 计算新区块的哈希
// 4. 将新区块添加到切片末尾
func (bc *Blockchain) AddBlock(data string) *Block {
	// 第1步：获取最后一个区块
	// len(bc.Blocks) - 1 是最后一个元素的索引
	prevBlock := bc.Blocks[len(bc.Blocks)-1]

	// 第2步：创建新区块
	newBlock := &Block{
		Index:        prevBlock.Index + 1,  // 索引递增
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: prevBlock.Hash,       // 链接！
		Nonce:        0,
	}

	// 第3步：计算哈希
	newBlock.Hash = newBlock.calculateHash()

	// 第4步：添加到区块链
	bc.Blocks = append(bc.Blocks, newBlock)

	return newBlock
}

// ============================================================================
// 区块链验证（完整性检查）
// ============================================================================

// Validate 验证区块链的完整性
//
// 验证什么？
// ==========
// 1. 每个区块的 Hash 是否正确（数据没有被篡改）
// 2. 每个 PreviousHash 是否正确指向（链条没有断裂）
//
// 为什么需要验证？
// ================
// 如果有人想篡改历史数据：
// 1. 修改区块1的数据 → Hash 变了
// 2. 区块2 的 PreviousHash 不再匹配 → 验证失败！
//
// 返回：
//   - true: 区块链完整有效
//   - false: 区块链被篡改或损坏
func (bc *Blockchain) Validate() bool {
	// 从第2个区块开始检查（索引1）
	// 创世区块没有 PreviousHash 需要验证
	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		// 验证1：检查当前区块的 Hash 是否正确
		// 重新计算哈希，看是否与存储的哈希一致
		currentHash := current.calculateHash()
		if current.Hash != currentHash {
			fmt.Printf("❌ 验证失败：区块 #%d 的哈希不匹配！\n", current.Index)
			fmt.Printf("   存储: %s\n", current.Hash)
			fmt.Printf("   计算: %s\n", currentHash)
			return false
		}

		// 验证2：检查 PreviousHash 是否正确
		// 前一个区块的 Hash 应该等于当前区块的 PreviousHash
		if current.PreviousHash != previous.Hash {
			fmt.Printf("❌ 验证失败：区块 #%d 的前哈希不匹配！\n", current.Index)
			fmt.Printf("   当前区块 PreviousHash: %s\n", current.PreviousHash)
			fmt.Printf("   前一个区块 Hash: %s\n", previous.Hash)
			return false
		}
	}

	return true
}

// ============================================================================
// 辅助方法
// ============================================================================

// PrintBlockchain 打印整个区块链
func (bc *Blockchain) PrintBlockchain() {
	fmt.Println("\n" + "=" + strings.Repeat("=", 49))
	fmt.Println("📊 区块链完整信息")
	fmt.Println("=" + strings.Repeat("=", 49))

	for _, block := range bc.Blocks {
		fmt.Printf("\n区块 #%d\n", block.Index)
		fmt.Printf("  时间: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("  数据: %s\n", block.Data)
		fmt.Printf("  前哈希: %s\n", block.PreviousHash)
		fmt.Printf("  哈希: %s\n", block.Hash)
	}
}

// GetLatestBlock 获取最后一个区块
func (bc *Blockchain) GetLatestBlock() *Block {
	return bc.Blocks[len(bc.Blocks)-1]
}

// GetLength 获取区块链长度
func (bc *Blockchain) GetLength() int {
	return len(bc.Blocks)
}

// ============================================================================
// 主函数：演示区块链的创建、添加和验证
// ============================================================================

func main() {
	fmt.Println("📚 第三课：区块链结构")
	fmt.Println("=" + strings.Repeat("=", 49))

	// -------------------------------------------------------------------------
	// 演示1：创建区块链
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：创建区块链（包含创世区块）")
	fmt.Println("----------------------------------------")

	blockchain := NewBlockchain()
	fmt.Printf("✅ 区块链创建成功！\n")
	fmt.Printf("   当前长度: %d 个区块\n", blockchain.GetLength())
	fmt.Printf("   创世区块哈希: %s\n", blockchain.Blocks[0].Hash[:16]+"...")

	// -------------------------------------------------------------------------
	// 演示2：添加新区块
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：添加交易记录")
	fmt.Println("----------------------------------------")

	transactions := []string{
		"Alice 转账给 Bob 50 BTC",
		"Bob 转账给 Charlie 30 BTC",
		"Charlie 转账给 Dave 20 BTC",
	}

	for _, tx := range transactions {
		block := blockchain.AddBlock(tx)
		fmt.Printf("✅ 添加区块 #%d: %s\n", block.Index, tx)
		fmt.Printf("   哈希: %s\n", block.Hash[:16]+"...")
	}

	// -------------------------------------------------------------------------
	// 演示3：验证区块链
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：验证区块链完整性")
	fmt.Println("----------------------------------------")

	valid := blockchain.Validate()
	if valid {
		fmt.Println("✅ 区块链验证通过！数据完整。")
	} else {
		fmt.Println("❌ 区块链验证失败！数据可能被篡改。")
	}

	// -------------------------------------------------------------------------
	// 演示4：篡改数据（展示验证的作用）
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：模拟数据篡改")
	fmt.Println("----------------------------------------")

	fmt.Println("😈 恶意攻击者试图修改区块 #1 的数据...")
	originalData := blockchain.Blocks[1].Data
	blockchain.Blocks[1].Data = "Alice 转账给 Bob 5000 BTC" // 篡改！
	fmt.Printf("   原始数据: %s\n", originalData)
	fmt.Printf("   篡改后: %s\n", blockchain.Blocks[1].Data)

	fmt.Println("\n🔍 验证区块链...")
	valid = blockchain.Validate()
	if valid {
		fmt.Println("⚠️  意外：验证通过了（不应该发生）")
	} else {
		fmt.Println("✅ 验证失败！篡改被检测到！")
	}

	// 恢复数据
	blockchain.Blocks[1].Data = originalData
	blockchain.Blocks[1].Hash = blockchain.Blocks[1].calculateHash()
	fmt.Println("\n🔄 恢复原始数据...")
	fmt.Println("✅ 区块链已恢复正常")

	// -------------------------------------------------------------------------
	// 演示5：打印完整区块链
	// -------------------------------------------------------------------------
	blockchain.PrintBlockchain()

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. 区块链是由多个区块按顺序连接的数据结构
	// 2. 创世区块是区块链的第一个区块
	// 3. 添加新区块时，PreviousHash 指向前一个区块的 Hash
	// 4. 可以通过验证哈希和链接来检测数据篡改
	// 5. 篡改任何区块的数据都会被验证发现
	//
	// ❓ 思考题：
	// 1. 如果篡改了区块1的数据，为什么只修复区块1的 Hash 不够？
	// 2. 在真实的比特币网络中，如何防止有人一直添加新区块？
	// 3. 空的区块链（只有创世区块）有效吗？
	//
	// 🔜 下一课：挖矿与工作量证明 - 学习如何控制区块添加的难度
	// ============================================================================
	fmt.Println("\n✅ 第三课完成！下一课将学习挖矿与工作量证明。")
}