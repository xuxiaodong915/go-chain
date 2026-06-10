// ============================================================================
// 第八课：P2P 网络 (Peer-to-Peer Network)
// ============================================================================
// 目标：理解区块链节点如何通过网络进行通信
// 前置知识：第一课 - 区块结构，第三课 - 区块链结构
// 学习时长：35分钟
//
// 什么是 P2P 网络？
// =================
// P2P (Peer-to-Peer) 是一种分布式网络架构，每个节点既是客户端也是服务器。
// 没有中心化的服务器，所有节点地位平等。
//
// 区块链为什么需要 P2P？
// =====================
// 1. 去中心化：没有单点故障
// 2. 数据同步：所有节点都有完整账本
// 3. 共识达成：节点间交换区块和交易
// 4. 抗审查：无法轻易关闭网络
//
// P2P 网络中的节点角色
// =====================
// 1. 全节点（Full Node）：存储完整区块链，验证所有交易
// 2. 轻节点（Light Node）：只存储区块头，依赖全节点
// 3. 矿工节点（Miner）：参与挖矿的全节点
// 4. SPV 节点：简化支付验证节点
//
// P2P 通信的内容
// ===============
// 1. 新交易广播：当有新交易时，广播给所有节点
// 2. 新区块广播：当有新区块时，广播给所有节点
// 3. 区块链同步：新节点加入时同步整个链
// 4. 对等节点发现：查找网络中的其他节点
// ============================================================================

//go:build lesson08

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// 区块基础结构（简化版，用于演示）
// ============================================================================

type Block struct {
	Index int64  `json:"index"`
	Data  string `json:"data"`
	Hash  string `json:"hash"`
}

type Blockchain struct {
	Blocks []*Block
	mu     sync.RWMutex
}

func NewBlockchain() *Blockchain {
	genesis := &Block{
		Index: 0,
		Data:  "创世区块",
		Hash:  "genesis_hash",
	}
	return &Blockchain{
		Blocks: []*Block{genesis},
	}
}

func (bc *Blockchain) AddBlock(data string) *Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Index: prevBlock.Index + 1,
		Data:  data,
		Hash:  fmt.Sprintf("hash_%d_%s", prevBlock.Index+1, data),
	}
	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock
}

func (bc *Blockchain) GetBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Blocks
}

// ============================================================================
// P2P 节点结构
// ============================================================================

// P2PNode P2P 节点
//
// 每个节点：
// 1. 有一个唯一 ID
// 2. 维护一个对等节点列表
// 3. 运行 HTTP 服务器接收消息
// 4. 可以广播消息给其他节点
type P2PNode struct {
	ID        string        // 节点 ID
	Address   string        // 节点地址（host:port）
	Peers     []string      // 对等节点列表
	Blockchain *Blockchain  // 本地区块链
	mu        sync.RWMutex  // 并发控制
	server    *http.Server  // HTTP 服务器
}

// NewP2PNode 创建新的 P2P 节点
//
// 参数：
//   - id: 节点 ID
//   - address: 监听地址（如 "localhost:8080"）
//   - blockchain: 区块链实例
func NewP2PNode(id, address string, blockchain *Blockchain) *P2PNode {
	return &P2PNode{
		ID:        id,
		Address:   address,
		Peers:     []string{},
		Blockchain: blockchain,
	}
}

// ============================================================================
// 节点管理方法
// ============================================================================

// AddPeer 添加对等节点
//
// 参数：
//   - peerAddress: 对等节点的地址
func (n *P2PNode) AddPeer(peerAddress string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 检查是否已存在
	for _, peer := range n.Peers {
		if peer == peerAddress {
			return
		}
	}

	n.Peers = append(n.Peers, peerAddress)
	fmt.Printf("🔗 节点 %s 连接到对等节点: %s\n", n.ID, peerAddress)
}

// RemovePeer 移除对等节点
func (n *P2PNode) RemovePeer(peerAddress string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for i, peer := range n.Peers {
		if peer == peerAddress {
			n.Peers = append(n.Peers[:i], n.Peers[i+1:]...)
			fmt.Printf("🔌 节点 %s 断开与 %s 的连接\n", n.ID, peerAddress)
			return
		}
	}
}

// GetPeers 获取所有对等节点
func (n *P2PNode) GetPeers() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Peers
}

// GetPeerCount 获取对等节点数量
func (n *P2PNode) GetPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.Peers)
}

// ============================================================================
// 消息广播
// ============================================================================

// Message 网络消息结构
//
// 所有节点间通信都使用这个统一的消息格式
type Message struct {
	Type      string                 `json:"type"`       // 消息类型：new_block, new_tx, etc.
	From      string                 `json:"from"`       // 发送方节点
	Timestamp int64                  `json:"timestamp"`  // 时间戳
	Payload   map[string]interface{} `json:"payload"`    // 消息内容
}

// Broadcast 广播消息给所有对等节点
//
// 参数：
//   - msg: 要广播的消息
//
// 广播是异步的，每个连接在一个 goroutine 中处理
func (n *P2PNode) Broadcast(msg Message) {
	n.mu.RLock()
	peers := make([]string, len(n.Peers))
	copy(peers, n.Peers)
	n.mu.RUnlock()

	if len(peers) == 0 {
		fmt.Printf("📢 节点 %s 没有对等节点，无法广播\n", n.ID)
		return
	}

	fmt.Printf("📡 节点 %s 广播消息类型 '%s' 给 %d 个节点\n",
		n.ID, msg.Type, len(peers))

	for _, peer := range peers {
		// 异步发送
		go n.sendToPeer(peer, msg)
	}
}

// sendToPeer 发送消息给指定节点
//
// 参数：
//   - peerAddress: 目标节点地址
//   - msg: 要发送的消息
func (n *P2PNode) sendToPeer(peerAddress string, msg Message) {
	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("❌ 序列化消息失败: %v\n", err)
		return
	}

	// 发送 HTTP POST 请求
	url := "http://" + peerAddress + "/receive"
	resp, err := http.Post(url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		fmt.Printf("❌ 发送给 %s 失败: %v\n", peerAddress, err)
		return
	}
	defer resp.Body.Close()

	// 读取响应（可选）
	io.Copy(io.Discard, resp.Body)

	fmt.Printf("  ✅ 消息已发送给 %s\n", peerAddress)
}

// ============================================================================
// HTTP 服务器
// ============================================================================

// Start 启动节点服务器
//
// 启动后会监听以下端点：
// - GET /          : 节点信息
// - GET /peers     : 对等节点列表
// - GET /blockchain: 本地区块链
// - POST /receive  : 接收其他节点广播的消息
func (n *P2PNode) Start() error {
	mux := http.NewServeMux()

	// GET / - 节点信息
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]interface{}{
			"id":       n.ID,
			"address":  n.Address,
			"peers":    len(n.Peers),
			"blocks":   len(n.Blockchain.Blocks),
			"uptime":   time.Since(time.Now()).Seconds(),
		}
		json.NewEncoder(w).Encode(info)
	})

	// GET /peers - 对等节点列表
	mux.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		n.mu.RLock()
		defer n.mu.RUnlock()
		json.NewEncoder(w).Encode(n.Peers)
	})

	// GET /blockchain - 本地区块链
	mux.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(n.Blockchain.GetBlocks())
	})

	// POST /receive - 接收消息
	mux.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 处理接收到的消息
		n.handleMessage(msg)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "received"})
	})

	// 启动服务器
	n.server = &http.Server{
		Addr:    n.Address,
		Handler: mux,
	}

	fmt.Printf("🌐 节点 %s 启动在 %s\n", n.ID, n.Address)
	return n.server.ListenAndServe()
}

// handleMessage 处理接收到的消息
func (n *P2PNode) handleMessage(msg Message) {
	fmt.Printf("📥 节点 %s 收到来自 %s 的消息: type=%s\n",
		n.ID, msg.From, msg.Type)

	switch msg.Type {
	case "new_block":
		// 处理新区块广播
		fmt.Println("   → 收到新区块广播")
		// 真实场景会验证并添加到本地链

	case "new_transaction":
		// 处理新交易广播
		fmt.Println("   → 收到新交易广播")
		// 真实场景会验证并添加到交易池

	case "ping":
		// 响应 ping 消息
		fmt.Println("   → 收到 ping，准备 pong")

	case "discover":
		// 节点发现
		fmt.Println("   → 收到节点发现请求")
	}
}

// ============================================================================
// 演示：节点发现
// ============================================================================

// DiscoverPeers 发现更多对等节点
//
// 向已知的对等节点询问他们的对等节点
func (n *P2PNode) DiscoverPeers() {
	peers := n.GetPeers()
	fmt.Printf("🔍 节点 %s 开始节点发现...\n", n.ID)

	for _, peer := range peers {
		resp, err := http.Get("http://" + peer + "/peers")
		if err != nil {
			continue
		}

		var theirPeers []string
		json.NewDecoder(resp.Body).Decode(&theirPeers)
		resp.Body.Close()

		// 添加他们知道但我们不知道的节点
		for _, theirPeer := range theirPeers {
			if theirPeer != n.Address {
				n.AddPeer(theirPeer)
			}
		}
	}

	fmt.Printf("✅ 节点发现完成，当前有 %d 个对等节点\n", n.GetPeerCount())
}

// SyncBlockchain 从对等节点同步区块链
//
// 简化版：从第一个对等节点获取完整区块链
func (n *P2PNode) SyncBlockchain() {
	peers := n.GetPeers()
	if len(peers) == 0 {
		fmt.Println("📭 没有对等节点可同步")
		return
	}

	// 从第一个对等节点获取区块链
	peer := peers[0]
	resp, err := http.Get("http://" + peer + "/blockchain")
	if err != nil {
		fmt.Printf("❌ 同步失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var blocks []*Block
	json.NewDecoder(resp.Body).Decode(&blocks)

	fmt.Printf("📥 从 %s 同步到 %d 个区块\n", peer, len(blocks))

	n.Blockchain.mu.Lock()
	n.Blockchain.Blocks = blocks
	n.Blockchain.mu.Unlock()
}

// ============================================================================
// 辅助函数
// ============================================================================

// PrintNodeInfo 打印节点信息
func (n *P2PNode) PrintNodeInfo() {
	fmt.Printf("\n📊 节点 %s 信息:\n", n.ID)
	fmt.Printf("   地址: %s\n", n.Address)
	fmt.Printf("   对等节点: %d 个\n", n.GetPeerCount())
	fmt.Printf("   区块数: %d 个\n", len(n.Blockchain.Blocks))
	if n.GetPeerCount() > 0 {
		fmt.Printf("   对等节点列表:\n")
		for _, peer := range n.GetPeers() {
			fmt.Printf("     - %s\n", peer)
		}
	}
}

// ============================================================================
// 主函数：演示 P2P 网络
// ============================================================================

func main() {
	fmt.Println("📚 第八课：P2P 网络")
	fmt.Println("=" + strings.Repeat("=", 49))

	// -------------------------------------------------------------------------
	// 演示1：创建多个节点
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：创建 P2P 节点")
	fmt.Println("----------------------------------------")

	node1 := NewP2PNode("Node1", "localhost:8080", NewBlockchain())
	node2 := NewP2PNode("Node2", "localhost:8081", NewBlockchain())
	node3 := NewP2PNode("Node3", "localhost:8082", NewBlockchain())

	fmt.Println("✅ 创建了 3 个节点")

	// -------------------------------------------------------------------------
	// 演示2：节点间连接
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：节点间连接")
	fmt.Println("----------------------------------------")

	// Node1 连接到 Node2
	node1.AddPeer("localhost:8081")

	// Node2 连接到 Node3
	node2.AddPeer("localhost:8082")

	// Node3 连接到 Node1（形成环状）
	node3.AddPeer("localhost:8080")

	node1.PrintNodeInfo()
	node2.PrintNodeInfo()
	node3.PrintNodeInfo()

	// -------------------------------------------------------------------------
	// 演示3：消息广播
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：消息广播")
	fmt.Println("----------------------------------------")

	// Node1 广播消息
	msg := Message{
		Type:      "new_block",
		From:      node1.ID,
		Timestamp: time.Now().Unix(),
		Payload: map[string]interface{}{
			"index": 1,
			"data":  "Alice 转账给 Bob 50 BTC",
		},
	}

	fmt.Println("\n📡 Node1 广播新区块消息...")
	node1.Broadcast(msg)

	time.Sleep(500 * time.Millisecond) // 等待消息传播

	// -------------------------------------------------------------------------
	// 演示4：节点发现
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：节点发现")
	fmt.Println("----------------------------------------")

	// Node2 发现更多节点
	node2.DiscoverPeers()
	node2.PrintNodeInfo()

	// -------------------------------------------------------------------------
	// 演示5：区块链同步
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示5：区块链同步")
	fmt.Println("----------------------------------------")

	// Node1 添加一些区块
	node1.Blockchain.AddBlock("交易1")
	node1.Blockchain.AddBlock("交易2")
	node1.Blockchain.AddBlock("交易3")

	fmt.Printf("Node1 现在有 %d 个区块\n", len(node1.Blockchain.Blocks))

	// Node3 从对等节点同步
	fmt.Println("\nNode3 开始同步...")
	node3.SyncBlockchain()

	fmt.Printf("Node3 同步后有 %d 个区块\n", len(node3.Blockchain.Blocks))

	// -------------------------------------------------------------------------
	// 演示6：网络拓扑
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示6：网络拓扑")
	fmt.Println("----------------------------------------")

	fmt.Println("当前网络结构:")
	fmt.Println("     Node1 (8080) ←→ Node2 (8081)")
	fmt.Println("       ↑                    ↓")
	fmt.Println("     Node3 (8082) ←←←←←←←←←←↓")
	fmt.Println("\n形成了一个环状网络，消息可以从任意节点传播到所有节点")

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. P2P 网络是去中心化的，节点地位平等
	// 2. 节点通过 HTTP API 通信
	// 3. 节点可以广播消息给所有对等节点
	// 4. 节点发现用于找到更多对等节点
	// 5. 区块链同步确保所有节点数据一致
	//
	// ❓ 思考题：
	// 1. 如果两个节点有不同的区块链，如何解决？（分叉处理）
	// 2. 如何防止恶意节点广播错误数据？
	// 3. 在真实的比特币网络中，节点使用什么协议通信？
	//
	// 🎉 课程总结：区块链学习路径
	// ============================================================================
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 恭喜！区块链课程全部完成！")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("\n📚 学习回顾:")
	fmt.Println("   第一课：区块结构 - 区块链的基本单元")
	fmt.Println("   第二课：哈希计算 - 数据的指纹")
	fmt.Println("   第三课：区块链结构 - 区块的链接")
	fmt.Println("   第四课：挖矿与工作量证明 - 控制区块添加")
	fmt.Println("   第五课：交易系统 - 价值转移")
	fmt.Println("   第六课：数字签名与钱包 - 身份认证")
	fmt.Println("   第七课：智能合约 - 可编程的区块链")
	fmt.Println("   第八课：P2P 网络 - 节点间通信")
	fmt.Println("\n🚀 下一步:")
	fmt.Println("   - 学习更多共识算法（PoS, DPoS, PBFT）")
	fmt.Println("   - 学习以太坊和 Solidity 智能合约开发")
	fmt.Println("   - 学习 Layer 2 扩展方案")
	fmt.Println("   - 实践：部署自己的测试网络")
	fmt.Println("\n" + strings.Repeat("=", 50))
}

// ============================================================================
// 注意：实际运行此程序需要取消注释启动服务器的代码
// ============================================================================
/*
func main() {
	// 实际运行时，需要在不同的终端或 goroutine 中启动多个节点

	node1 := NewP2PNode("Node1", "localhost:8080", NewBlockchain())
	node2 := NewP2PNode("Node2", "localhost:8081", NewBlockchain())

	// 连接节点
	node1.AddPeer("localhost:8081")
	node2.AddPeer("localhost:8080")

	// 在不同的 goroutine 中启动服务器
	go func() {
		if err := node1.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := node2.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// 等待服务器启动
	time.Sleep(time.Second)

	// 广播消息
	msg := Message{
		Type:      "ping",
		From:      node1.ID,
		Timestamp: time.Now().Unix(),
		Payload:   nil,
	}
	node1.Broadcast(msg)

	// 保持程序运行
	select {}
}
*/