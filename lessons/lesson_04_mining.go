// ============================================================================
// 第四课：挖矿与工作量证明 (Mining & Proof of Work)
// ============================================================================
// 目标：理解工作量证明机制，以及挖矿的概念
// 前置知识：第一课 - 区块结构，第二课 - 哈希计算
// 学习时长：30分钟
//
// 什么是挖矿？
// ============
// 在区块链中，"挖矿"不是真的挖矿，而是指找到一个满足特定条件的哈希值。
// 这个条件通常是：哈希值必须以若干个0开头。
//
// 什么是工作量证明（Proof of Work）？
// ====================================
// 工作量证明是一种共识机制，要求参与者（矿工）完成一定量的"工作"
// （计算哈希）才能获得添加区块的权利。
//
// 为什么要挖矿？
// ==============
// 1. 防止随意添加区块：需要付出计算成本
// 2. 控制区块生成速度：难度越高，挖矿越慢
// 3. 防止双花攻击：篡改历史需要重新挖链上所有区块
//
// 挖矿的过程
// ===========
// 1. 准备区块数据
// 2. 设置目标：比如哈希必须以 "000" 开头
// 3. 不断改变 Nonce 值，重新计算哈希
// 4. 找到符合条件的哈希时，停止计算
// 5. 将该区块添加到区块链
//
// 什么是 Nonce？
// =============
// Nonce (Number used once) 是一个只使用一次的随机数。
// 在挖矿过程中，我们不断改变 Nonce 的值，直到找到符合条件的哈希。
// ============================================================================

//go:build lesson04

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Block 区块结构
type Block struct {
	Index        int64
	Timestamp    int64
	Data         string
	PreviousHash string
	Hash         string
	Nonce        int64 // 挖矿时不断改变的值
}

func (b *Block) calculateHash() string {
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.Data, b.PreviousHash, b.Nonce)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================================================
// 核心方法：mineBlock（挖矿）
// ============================================================================

// mineBlock 挖矿：找到满足条件的哈希
//
// 参数：
//   - difficulty: 难度级别
//                  difficulty = 1 表示哈希以 1 个 0 开头
//                  difficulty = 2 表示哈希以 2 个 0 开头
//                  难度每增加 1，平均计算次数翻倍！
//
// 难度与计算次数的关系：
// =====================
// 难度 1: 平均需要尝试 16 次（1/16 的概率，因为十六进制有 16 个字符）
// 难度 2: 平均需要尝试 256 次
// 难度 3: 平均需要尝试 4096 次
// 难度 4: 平均需要尝试 65536 次
// 难度 5: 平均需要尝试 1,048,576 次
// ...
// 难度 n: 平均需要尝试 16^n 次
//
// 真实比特币的难度：
// ==================
// 比特币的难度是动态调整的，目标是每 10 分钟产生一个区块。
// 目前的难度使得哈希要以约 70 个 0 开头！
func (b *Block) mineBlock(difficulty int) {
	// 第1步：根据难度确定目标字符串
	// 比如难度为 3，目标就是 "000"
	target := strings.Repeat("0", difficulty)

	fmt.Printf("⛏️  开始挖矿区块 #%d，难度: %d (目标: 哈希以 '%s' 开头)\n",
		b.Index, difficulty, target)

	startTime := time.Now()
	attempts := 0

	// 第2步：不断尝试不同的 Nonce 值
	for {
		attempts++
		hash := b.calculateHash()

		// 第3步：检查哈希是否满足条件
		// hash[:difficulty] 获取哈希的前 difficulty 个字符
		if hash[:difficulty] == target {
			b.Hash = hash
			duration := time.Since(startTime)

			fmt.Printf("✅ 挖矿成功！\n")
			fmt.Printf("   哈希: %s\n", hash)
			fmt.Printf("   Nonce: %d\n", b.Nonce)
			fmt.Printf("   尝试次数: %d\n", attempts)
			fmt.Printf("   耗时: %v\n", duration)
			return
		}

		// 不满足条件，递增 Nonce 再试
		b.Nonce++

		// 安全阀：防止无限循环
		if attempts > 10000000 {
			fmt.Printf("❌ 挖矿失败：尝试次数过多，请降低难度\n")
			return
		}
	}
}

// NewBlock 创建新区块并立即挖矿
func NewBlock(index int64, data string, previousHash string, difficulty int) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: previousHash,
		Nonce:        0, // 从 0 开始尝试
	}

	// 挖矿！
	block.mineBlock(difficulty)

	return block
}

// ============================================================================
// 主函数：演示挖矿过程
// ============================================================================

func main() {
	fmt.Println("📚 第四课：挖矿与工作量证明")
	fmt.Println("=" + strings.Repeat("=", 49))

	// -------------------------------------------------------------------------
	// 演示1：低难度挖矿（难度1）
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：难度为 1 的挖矿")
	fmt.Println("----------------------------------------")

	block1 := NewBlock(
		1,
		"Alice 转账给 Bob 50 BTC",
		"genesis_hash_placeholder",
		1, // 难度为 1
	)

	// -------------------------------------------------------------------------
	// 演示2：中等难度挖矿（难度2）
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：难度为 2 的挖矿")
	fmt.Println("----------------------------------------")

	block2 := NewBlock(
		2,
		"Bob 转账给 Charlie 30 BTC",
		block1.Hash,
		2, // 难度为 2
	)

	// -------------------------------------------------------------------------
	// 演示3：高难度挖矿（难度3）- 注意观察尝试次数
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：难度为 3 的挖矿")
	fmt.Println("----------------------------------------")

	block3 := NewBlock(
		3,
		"Charlie 转账给 Dave 20 BTC",
		block2.Hash,
		3, // 难度为 3
	)

	// -------------------------------------------------------------------------
	// 演示4：展示所有挖出的区块
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：所有挖出的区块")
	fmt.Println("----------------------------------------")

	blocks := []*Block{block1, block2, block3}
	for _, block := range blocks {
		fmt.Printf("\n区块 #%d\n", block.Index)
		fmt.Printf("  数据: %s\n", block.Data)
		fmt.Printf("  哈希: %s\n", block.Hash)
		fmt.Printf("  前哈希: %s\n", block.PreviousHash)
		fmt.Printf("  Nonce: %d\n", block.Nonce)
	}

	// -------------------------------------------------------------------------
	// 演示5：验证哈希确实满足条件
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示5：验证哈希满足难度条件")
	fmt.Println("----------------------------------------")

	difficulties := []int{1, 2, 3}
	for _, diff := range difficulties {
		target := strings.Repeat("0", diff)
		fmt.Printf("难度 %d: 哈希以 '%s' 开头？\n", diff, target)

		var targetBlock *Block
		switch diff {
		case 1:
			targetBlock = block1
		case 2:
			targetBlock = block2
		case 3:
			targetBlock = block3
		}

		hashPrefix := targetBlock.Hash[:diff]
		if hashPrefix == target {
			fmt.Printf("  ✅ 是的！%s == %s\n", hashPrefix, target)
		} else {
			fmt.Printf("  ❌ 不是！%s != %s\n", hashPrefix, target)
		}
	}

	// -------------------------------------------------------------------------
	// 演示6：模拟竞争挖矿
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示6：模拟矿工竞争")
	fmt.Println("----------------------------------------")

	fmt.Println("🎮 3个矿工同时挖同一个区块...")
	fmt.Println("   矿工A: 从 Nonce=0 开始")
	fmt.Println("   矿工B: 从 Nonce=1000 开始")
	fmt.Println("   矿工C: 从 Nonce=2000 开始")

	// 这是一个简单的模拟，真实情况是每个矿工有不同的数据（地址、时间戳）
	type Miner struct {
		Name      string
		StartNonce int64
		Nonce     int64
		Hash      string
		Found     bool
		Attempts  int64
	}

	miners := []*Miner{
		{Name: "矿工A", StartNonce: 0},
		{Name: "矿工B", StartNonce: 1000},
		{Name: "矿工C", StartNonce: 2000},
	}

	target := strings.Repeat("0", 2) // 难度2

	// 模拟挖矿（简化版）
	for {
		allDone := true
		for _, miner := range miners {
			if !miner.Found {
				allDone = false
				miner.Nonce = miner.StartNonce + miner.Attempts
				miner.Attempts++

				// 计算这个 Nonce 对应的哈希
				record := fmt.Sprintf("%d%d%s%s%d",
					99, time.Now().Unix(), "竞争区块", "prev_hash", miner.Nonce)
				h := sha256.New()
				h.Write([]byte(record))
				hash := hex.EncodeToString(h.Sum(nil))

				// 检查是否满足条件
				if hash[:2] == target {
					miner.Hash = hash
					miner.Found = true
					fmt.Printf("🎉 %s 找到有效区块！\n", miner.Name)
					fmt.Printf("   Nonce: %d, 尝试次数: %d\n", miner.Nonce, miner.Attempts)
					fmt.Printf("   哈希: %s\n", hash[:16]+"...")
					break
				}
			}
		}
		if allDone {
			break
		}
	}

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. 挖矿是寻找满足特定条件（前导0）的哈希值的过程
	// 2. 工作量证明通过增加计算成本防止恶意行为
	// 3. Nonce 是挖矿时不断改变的随机数
	// 4. 难度每增加 1，平均计算次数翻倍
	// 5. 多个矿工可以竞争挖矿，先找到的获胜
	//
	// ❓ 思考题：
	// 1. 为什么难度越高，挖矿越困难？
	// 2. 如果所有人的计算机速度都变快了，会发生什么？（比特币如何应对？）
	// 3. 挖矿成功后，矿工会得到什么奖励？
	//
	// 🔜 下一课：交易系统 - 学习如何处理真实的交易记录
	// ============================================================================
	fmt.Println("\n✅ 第四课完成！下一课将学习交易系统。")
}