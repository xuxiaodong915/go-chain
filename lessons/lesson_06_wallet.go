// ============================================================================
// 第六课：数字签名与钱包 (Digital Signature & Wallet)
// ============================================================================
// 目标：理解公私钥加密和数字签名，实现钱包功能
// 前置知识：第一课 - 区块结构，第五课 - 交易系统
// 学习时长：35分钟
//
// 什么是数字签名？
// ================
// 数字签名是手写签名的电子版，它能证明：
// 1. 消息确实来自发送方（身份认证）
// 2. 消息没有被篡改（完整性）
// 3. 发送方无法否认发送过这条消息（不可抵赖）
//
// 公私钥加密
// ============
// 每个钱包有一对密钥：
// - 私钥（Private Key）：只有钱包持有者知道，绝对保密！用于签名
// - 公钥（Public Key）：可以公开分享，用于验证签名
//
// 数字签名的过程
// ================
// 1. 发送方用私钥对交易签名
// 2. 将交易和签名一起发送
// 3. 接收方用发送方的公钥验证签名
// 4. 如果验证通过，说明交易确实来自发送方
//
// 为什么无法伪造？
// ================
// 私钥和公钥是数学相关的，但无法从公钥推导出私钥。
// 只有持有私钥的人才能生成正确的签名。
// ============================================================================

//go:build lesson06

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ============================================================================
// 数字签名相关类型和函数
// ============================================================================

// Signature 数字签名结构（包含 R 和 S 两个大整数）
type Signature struct {
	R *big.Int
	S *big.Int
}

// String 将签名转换为十六进制字符串
func (sig *Signature) String() string {
	return hex.EncodeToString(append(sig.R.Bytes(), sig.S.Bytes()...))
}

// ParseSignature 从字符串解析签名
func ParseSignature(sigStr string) (*Signature, error) {
	data, err := hex.DecodeString(sigStr)
	if err != nil {
		return nil, err
	}

	// ECDSA P-256 生成的签名是 64 字节
	// 前 32 字节是 R，后 32 字节是 S
	if len(data) != 64 {
		return nil, fmt.Errorf("无效的签名长度")
	}

	return &Signature{
		R: new(big.Int).SetBytes(data[:32]),
		S: new(big.Int).SetBytes(data[32:]),
	}, nil
}

// ============================================================================
// 钱包结构
// ============================================================================

// Wallet 钱包结构
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey // 私钥（保密！）
	PublicKey  *ecdsa.PublicKey  // 公钥（可公开）
	Address    string           // 地址（从公钥生成）
}

// NewWallet 创建新钱包
//
// 流程：
// 1. 使用 ECDSA P-256 曲线生成密钥对
// 2. 从公钥生成钱包地址
func NewWallet() *Wallet {
	// 第1步：生成密钥对
	// elliptic.P256() 是一种椭圆曲线算法
	// rand.Reader 是随机数源
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("生成密钥失败: %v", err))
	}

	// 第2步：从公钥生成地址
	address := generateAddress(&privateKey.PublicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Address:    address,
	}
}

// generateAddress 从公钥生成钱包地址
//
// 在真实区块链中，地址生成方式更复杂：
// - 比特币：公钥 -> SHA256 -> RIPEMD160 -> Base58Check 编码
// - 以太坊：公钥 -> Keccak256 -> 取后20字节 -> 0x 前缀
//
// 这里简化为：公钥的 SHA256 哈希
func generateAddress(pubKey *ecdsa.PublicKey) string {
	// 将公钥转换为字节
	// elliptic.Marshal 返回公钥的编码形式
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), pubKey.X, pubKey.Y)

	// 计算 SHA256 哈希
	h := sha256.New()
	h.Write(pubKeyBytes)
	address := hex.EncodeToString(h.Sum(nil))

	return address
}

// ============================================================================
// 签名方法
// ============================================================================

// Sign 对数据（交易）进行签名
//
// 签名步骤：
// 1. 对数据进行哈希
// 2. 用私钥对哈希值签名
// 3. 返回签名（R, S）
func (w *Wallet) Sign(data []byte) (*Signature, error) {
	// 第1步：计算数据的哈希
	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)

	// 第2步：用私钥签名
	// rand.Reader 用于生成签名的随机数（k值）
	// 注意：每次签名结果不同，但都能验证通过！
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %v", err)
	}

	return &Signature{R: r, S: s}, nil
}

// ============================================================================
// 验证方法
// ============================================================================

// Verify 验证签名是否有效
//
// 参数：
//   - pubKey: 公钥
//   - data: 原始数据
//   - signature: 签名
//
// 返回：
//   - bool: 签名是否有效
func Verify(pubKey *ecdsa.PublicKey, data []byte, signature *Signature) bool {
	// 第1步：计算数据的哈希
	h := sha256.New()
	h.Write(data)
	hash := h.Sum(nil)

	// 第2步：用公钥验证签名
	// 注意：需要同时传入 R 和 S
	valid := ecdsa.Verify(pubKey, hash, signature.R, signature.S)

	return valid
}

// VerifyString 验证字符串形式的签名
func VerifyString(pubKeyStr string, data []byte, sigStr string) bool {
	// 从字符串解析公钥（简化版）
	pubKeyBytes, _ := hex.DecodeString(pubKeyStr)
	x := new(big.Int).SetBytes(pubKeyBytes[:32])
	y := new(big.Int).SetBytes(pubKeyBytes[32:64])
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// 解析签名
	sig, err := ParseSignature(sigStr)
	if err != nil {
		return false
	}

	return Verify(pubKey, data, sig)
}

// ============================================================================
// 签名交易
// ============================================================================

// Transaction 交易结构（带签名）
type Transaction struct {
	From      string  // 发送方地址
	To        string  // 接收方地址
	Amount    float64 // 金额
	Signature string  // 数字签名
	Timestamp int64   // 时间戳
}

// SignTransaction 用钱包签名交易
//
// 流程：
// 1. 创建未签名交易
// 2. 将交易数据序列化为字节
// 3. 用私钥签名
// 4. 将签名加入交易
func (w *Wallet) SignTransaction(tx *Transaction) error {
	// 第1步：创建待签名的数据（不包含签名本身）
	txData := map[string]interface{}{
		"from":      tx.From,
		"to":        tx.To,
		"amount":    tx.Amount,
		"timestamp": tx.Timestamp,
	}

	// 第2步：序列化为 JSON
	data, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	// 第3步：签名
	sig, err := w.Sign(data)
	if err != nil {
		return err
	}

	// 第4步：加入签名
	tx.Signature = sig.String()

	return nil
}

// ============================================================================
// 主函数：演示数字签名
// ============================================================================

func main() {
	fmt.Println("📚 第六课：数字签名与钱包")
	fmt.Println("=" + strings.Repeat("=", 49))

	// -------------------------------------------------------------------------
	// 演示1：创建钱包
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：创建钱包")
	fmt.Println("----------------------------------------")

	alice := NewWallet()
	bob := NewWallet()

	fmt.Printf("✅ Alice 的钱包:\n")
	fmt.Printf("   地址: %s\n", alice.Address[:24]+"...")
	fmt.Printf("   公钥: %s\n", encodePublicKey(alice.PublicKey)[:24]+"...")
	fmt.Printf("   私钥: [保密]\n")

	fmt.Printf("\n✅ Bob 的钱包:\n")
	fmt.Printf("   地址: %s\n", bob.Address[:24]+"...")
	fmt.Printf("   公钥: %s\n", encodePublicKey(bob.PublicKey)[:24]+"...")
	fmt.Printf("   私钥: [保密]\n")

	// -------------------------------------------------------------------------
	// 演示2：基本签名和验证
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：数字签名的基本使用")
	fmt.Println("----------------------------------------")

	message := []byte("Hello, Bob! This is Alice.")

	// Alice 签名
	fmt.Println("Alice 正在签名消息...")
	sig, err := alice.Sign(message)
	if err != nil {
		fmt.Printf("签名失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 签名成功: %s\n", sig.String()[:24]+"...")

	// 验证签名
	fmt.Println("\nBob 正在验证签名...")
	valid := Verify(alice.PublicKey, message, sig)
	if valid {
		fmt.Println("✅ 签名验证通过！消息确实来自 Alice。")
	} else {
		fmt.Println("❌ 签名验证失败！")
	}

	// -------------------------------------------------------------------------
	// 演示3：篡改消息会被检测
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：消息篡改检测")
	fmt.Println("----------------------------------------")

	tamperedMessage := []byte("Hello, Bob! This is not Alice.")
	fmt.Printf("原始消息: %s\n", string(message))
	fmt.Printf("篡改消息: %s\n", string(tamperedMessage))

	fmt.Println("\n用篡改后的消息验证签名...")
	valid = Verify(alice.PublicKey, tamperedMessage, sig)
	if valid {
		fmt.Println("⚠️  签名验证通过（不应该发生！）")
	} else {
		fmt.Println("✅ 签名验证失败！检测到消息被篡改。")
	}

	// -------------------------------------------------------------------------
	// 演示4：签名交易
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：签名交易")
	fmt.Println("----------------------------------------")

	// Alice 创建转账交易
	tx := &Transaction{
		From:      alice.Address,
		To:        bob.Address,
		Amount:    50,
		Timestamp: time.Now().Unix(),
	}

	fmt.Printf("交易: Alice -> %.2f -> Bob\n", tx.Amount)
	fmt.Printf("Alice 正在签名交易...\n")

	// Alice 签名交易
	err = alice.SignTransaction(tx)
	if err != nil {
		fmt.Printf("签名失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 交易已签名\n")
	fmt.Printf("   签名: %s\n", tx.Signature[:24]+"...")

	// -------------------------------------------------------------------------
	// 演示5：验证交易签名
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示5：验证交易签名")
	fmt.Println("----------------------------------------")

	// 准备验证数据
	txData := map[string]interface{}{
		"from":      tx.From,
		"to":        tx.To,
		"amount":    tx.Amount,
		"timestamp": tx.Timestamp,
	}
	data, _ := json.Marshal(txData)

	// 解析签名
	sig, err = ParseSignature(tx.Signature)
	if err != nil {
		fmt.Printf("解析签名失败: %v\n", err)
		return
	}

	// 验证
	valid = Verify(alice.PublicKey, data, sig)
	if valid {
		fmt.Println("✅ 交易签名验证通过！")
	} else {
		fmt.Println("❌ 交易签名验证失败！")
	}

	// -------------------------------------------------------------------------
	// 演示6：私钥保密性
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示6：私钥的重要性")
	fmt.Println("----------------------------------------")

	fmt.Println("如果私钥泄露会发生什么？")
	fmt.Println("→ 攻击者可以用私钥签名任何交易，转走你的钱！")
	fmt.Println("→ 所以永远不要分享私钥！")
	fmt.Println("→ 将私钥安全存储（如硬件钱包、助记词）")

	// 打印私钥（仅用于演示，真实情况绝不应该这样做！）
	fmt.Printf("\n⚠️  Alice 的私钥（仅演示）: %s\n", encodePrivateKey(alice.PrivateKey))

	// 有人拿到私钥后，可以创建假签名
	fmt.Println("\n🔓 攻击者拿到 Alice 的私钥后...")

	// 攻击者用 Alice 的私钥签名交易
	attackerTx := &Transaction{
		From:      alice.Address,
		To:        "attacker_address",
		Amount:    1000,
		Timestamp: time.Now().Unix(),
	}

	err = alice.SignTransaction(attackerTx) // 攻击者用 Alice 的钱包
	if err == nil {
		fmt.Println("❌ 攻击者成功签名了交易！")
		fmt.Printf("   交易: Alice -> %.2f -> 攻击者\n", attackerTx.Amount)
		fmt.Println("   → 这就是为什么私钥必须保密！")
	}

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. 钱包包含公钥和私钥对
	// 2. 私钥用于签名，公钥用于验证
	// 3. 数字签名保证消息的真实性和完整性
	// 4. 篡改消息会导致签名验证失败
	// 5. 私钥泄露意味着钱包被盗
	//
	// ❓ 思考题：
	// 1. 为什么每次签名结果不同，但都能验证通过？
	// 2. 如果丢失了私钥，能找回钱包吗？
	// 3. 多重签名（Multi-sig）是什么意思？
	//
	// 🔜 下一课：智能合约 - 学习可编程的区块链
	// ============================================================================
	fmt.Println("\n✅ 第六课完成！下一课将学习智能合约。")
}

// ============================================================================
// 辅助函数
// ============================================================================

func encodePublicKey(pubKey *ecdsa.PublicKey) string {
	return hex.EncodeToString(append(pubKey.X.Bytes(), pubKey.Y.Bytes()...))
}

func encodePrivateKey(privKey *ecdsa.PrivateKey) string {
	return hex.EncodeToString(append(append(privKey.D.Bytes(), privKey.PublicKey.X.Bytes()...), privKey.PublicKey.Y.Bytes()...))
}