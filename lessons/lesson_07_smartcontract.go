// ============================================================================
// 第七课：智能合约 (Smart Contract)
// ============================================================================
// 目标：理解智能合约的概念，实现简单的合约执行引擎
// 前置知识：第五课 - 交易系统，第六课 - 数字签名
// 学习时长：40分钟
//
// 什么是智能合约？
// ================
// 智能合约是运行在区块链上的程序，它：
// 1. 预定义了一套规则
// 2. 当满足条件时自动执行
// 3. 结果不可篡改
//
// 智能合约的特点
// ===============
// 1. 自动执行：不需要人工干预
// 2. 去中心化：运行在所有节点上
// 3. 不可篡改：一旦部署无法修改
// 4. 透明公开：代码和状态都是公开的
//
// 智能合约的应用场景
// ===================
// - 代币发行（ERC-20）
// - 去中心化交易（DEX）
// - NFT（ERC-721）
// - DeFi（去中心化金融）
// - 投票系统
// - 供应链追溯
//
// 智能合约的生命周期
// ===================
// 1. 开发：编写合约代码
// 2. 编译：将代码编译为字节码
// 3. 部署：发送部署交易，合约上链
// 4. 调用：发送交易触发合约函数
// ============================================================================

//go:build lesson07

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ============================================================================
// 智能合约接口
// ============================================================================

// Contract 合约接口
//
// 所有智能合约都需要实现这个接口
// 这样我们的区块链就可以支持任意类型的合约
type Contract interface {
	// Execute 执行合约方法
	// ctx: 执行上下文，包含存储和调用者信息
	// method: 要调用的方法名
	// params: 方法参数
	Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{}

	// GetName 获取合约名称
	GetName() string
}

// ContractContext 合约执行上下文
//
// 合约执行时需要的环境信息：
// - Storage: 合约的存储空间（状态）
// - Caller: 调用者的地址
// - BlockNumber: 当前区块号
type ContractContext struct {
	Storage     map[string]string // 合约状态存储
	Caller      string            // 调用者地址
	BlockNumber int64             // 当前区块号
}

// ============================================================================
// 示例1：简单计数器合约
// ============================================================================

// CounterContract 计数器合约
//
// 功能：
// - increment(): 计数加1
// - decrement(): 计数减1
// - get(): 获取当前值
// - reset(): 重置为0
type CounterContract struct {
	name string
}

func NewCounterContract() *CounterContract {
	return &CounterContract{
		name: "Counter",
	}
}

func (c *CounterContract) GetName() string {
	return c.name
}

func (c *CounterContract) Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{} {
	switch method {
	case "increment":
		// 获取当前值
		current, _ := strconv.Atoi(ctx.Storage["count"])
		// 加1
		newValue := current + 1
		// 保存
		ctx.Storage["count"] = strconv.Itoa(newValue)
		return map[string]interface{}{
			"success": true,
			"oldValue": current,
			"newValue": newValue,
		}

	case "decrement":
		current, _ := strconv.Atoi(ctx.Storage["count"])
		newValue := current - 1
		ctx.Storage["count"] = strconv.Itoa(newValue)
		return map[string]interface{}{
			"success": true,
			"oldValue": current,
			"newValue": newValue,
		}

	case "get":
		count, _ := strconv.Atoi(ctx.Storage["count"])
		return map[string]interface{}{
			"count": count,
		}

	case "reset":
		ctx.Storage["count"] = "0"
		return map[string]interface{}{
			"success": true,
			"message": "Counter reset to 0",
		}

	default:
		return map[string]interface{}{
			"error": "Unknown method: " + method,
		}
	}
}

// ============================================================================
// 示例2：代币合约 (简化版 ERC-20)
// ============================================================================

// TokenContract 代币合约
//
// 功能：
// - mint(address, amount): 铸造代币（只有所有者可以调用）
// - transfer(from, to, amount): 转账
// - balanceOf(address): 查询余额
// - totalSupply(): 查询总供应量
type TokenContract struct {
	name        string
	symbol      string
	totalSupply string
	owner       string
}

func NewTokenContract(name, symbol, owner string) *TokenContract {
	return &TokenContract{
		name:   name,
		symbol: symbol,
		owner:  owner,
	}
}

func (t *TokenContract) GetName() string {
	return t.name
}

func (t *TokenContract) Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{} {
	switch method {
	case "totalSupply":
		supply, _ := strconv.ParseFloat(t.totalSupply, 64)
		return map[string]interface{}{
			"totalSupply": supply,
			"symbol": t.symbol,
		}

	case "balanceOf":
		address := params["address"].(string)
		balance, exists := ctx.Storage["balance_"+address]
		if !exists {
			return map[string]interface{}{
				"address": address,
				"balance": 0,
			}
		}
		b, _ := strconv.ParseFloat(balance, 64)
		return map[string]interface{}{
			"address": address,
			"balance": b,
		}

	case "mint":
		// 验证：只有合约所有者可以铸造
		if ctx.Caller != t.owner {
			return map[string]interface{}{
				"error": "Only contract owner can mint",
			}
		}

		address := params["address"].(string)
		amount := params["amount"].(string)

		// 更新余额
		currentBalance, _ := strconv.ParseFloat(ctx.Storage["balance_"+address], 64)
		amountFloat, _ := strconv.ParseFloat(amount, 64)
		newBalance := currentBalance + amountFloat

		ctx.Storage["balance_"+address] = strconv.FormatFloat(newBalance, 'f', -1, 64)

		// 更新总供应量
		total, _ := strconv.ParseFloat(t.totalSupply, 64)
		t.totalSupply = strconv.FormatFloat(total+amountFloat, 'f', -1, 64)

		return map[string]interface{}{
			"success": true,
			"address": address,
			"minted": amount,
			"newBalance": newBalance,
		}

	case "transfer":
		from := params["from"].(string)
		to := params["to"].(string)
		amount := params["amount"].(string)

		// 获取余额
		fromBalance, _ := strconv.ParseFloat(ctx.Storage["balance_"+from], 64)
		toBalance, _ := strconv.ParseFloat(ctx.Storage["balance_"+to], 64)
		amountFloat, _ := strconv.ParseFloat(amount, 64)

		// 验证：余额充足
		if fromBalance < amountFloat {
			return map[string]interface{}{
				"error": "Insufficient balance",
				"fromBalance": fromBalance,
				"required": amountFloat,
			}
		}

		// 执行转账
		ctx.Storage["balance_"+from] = strconv.FormatFloat(fromBalance-amountFloat, 'f', -1, 64)
		ctx.Storage["balance_"+to] = strconv.FormatFloat(toBalance+amountFloat, 'f', -1, 64)

		return map[string]interface{}{
			"success": true,
			"from": from,
			"to": to,
			"amount": amount,
			"fromBalance": fromBalance - amountFloat,
			"toBalance": toBalance + amountFloat,
		}

	case "approve":
		// 授权第三方使用代币（简化版，只记录授权）
		owner := params["owner"].(string)
		spender := params["spender"].(string)
		amount := params["amount"].(string)

		key := fmt.Sprintf("allowance_%s_%s", owner, spender)
		ctx.Storage[key] = amount

		return map[string]interface{}{
			"success": true,
			"owner": owner,
			"spender": spender,
			"allowance": amount,
		}

	default:
		return map[string]interface{}{
			"error": "Unknown method: " + method,
		}
	}
}

// ============================================================================
// 示例3：投票合约
// ============================================================================

// VotingContract 投票合约
//
// 功能：
// - addCandidate(name): 添加候选人
// - vote(candidate): 投票
// - getResults(): 获取投票结果
// - hasVoted(address): 检查是否已投票
type VotingContract struct {
	name string
}

func NewVotingContract() *VotingContract {
	return &VotingContract{
		name: "Voting",
	}
}

func (t *VotingContract) GetName() string {
	return t.name
}

func (t *VotingContract) Execute(ctx *ContractContext, method string, params map[string]interface{}) interface{} {
	switch method {
	case "addCandidate":
		name := params["name"].(string)
		key := "candidate_" + name

		// 检查候选人是否已存在
		if _, exists := ctx.Storage[key]; exists {
			return map[string]interface{}{
				"error": "Candidate already exists",
			}
		}

		// 添加候选人
		ctx.Storage[key] = "0" // 初始票数为0

		return map[string]interface{}{
			"success": true,
			"candidate": name,
		}

	case "vote":
		candidate := params["candidate"].(string)
		caller := ctx.Caller

		// 检查候选人是否存在
		key := "candidate_" + candidate
		if _, exists := ctx.Storage[key]; !exists {
			return map[string]interface{}{
				"error": "Candidate does not exist",
			}
		}

		// 检查是否已投票
		votedKey := "voted_" + caller
		if _, hasVoted := ctx.Storage[votedKey]; hasVoted {
			return map[string]interface{}{
				"error": "Already voted",
			}
		}

		// 记录已投票
		ctx.Storage[votedKey] = "true"

		// 增加票数
		votes, _ := strconv.Atoi(ctx.Storage[key])
		ctx.Storage[key] = strconv.Itoa(votes + 1)

		return map[string]interface{}{
			"success": true,
			"candidate": candidate,
			"totalVotes": votes + 1,
		}

	case "hasVoted":
		address := params["address"].(string)
		votedKey := "voted_" + address
		_, hasVoted := ctx.Storage[votedKey]

		return map[string]interface{}{
			"address": address,
			"hasVoted": hasVoted,
		}

	case "getResults":
		results := make(map[string]int)
		for key, value := range ctx.Storage {
			if strings.HasPrefix(key, "candidate_") {
				name := strings.TrimPrefix(key, "candidate_")
				votes, _ := strconv.Atoi(value)
				results[name] = votes
			}
		}
		return map[string]interface{}{
			"results": results,
		}

	default:
		return map[string]interface{}{
			"error": "Unknown method: " + method,
		}
	}
}

// ============================================================================
// 合约注册表
// ============================================================================

// ContractRegistry 合约注册表
//
// 用于管理所有已部署的合约
type ContractRegistry struct {
	contracts map[string]Contract          // 合约实例
	storages  map[string]map[string]string // 每个合约的存储空间
}

func NewContractRegistry() *ContractRegistry {
	return &ContractRegistry{
		contracts: make(map[string]Contract),
		storages:  make(map[string]map[string]string),
	}
}

// Deploy 部署合约
//
// 参数：
//   - name: 合约名称
//   - contract: 合约实例
//
// 返回：
//   - bool: 是否部署成功
func (cr *ContractRegistry) Deploy(name string, contract Contract) bool {
	if _, exists := cr.contracts[name]; exists {
		return false
	}

	cr.contracts[name] = contract
	cr.storages[name] = make(map[string]string)
	return true
}

// Execute 执行合约方法
//
// 参数：
//   - name: 合约名称
//   - caller: 调用者地址
//   - method: 方法名
//   - params: 方法参数
//
// 返回：
//   - interface{}: 执行结果
func (cr *ContractRegistry) Execute(name string, caller string, method string, params map[string]interface{}) interface{} {
	contract, exists := cr.contracts[name]
	if !exists {
		return map[string]interface{}{
			"error": "Contract not found: " + name,
		}
	}

	ctx := &ContractContext{
		Storage:     cr.storages[name],
		Caller:      caller,
		BlockNumber: 0,
	}

	return contract.Execute(ctx, method, params)
}

// GetStorage 获取合约的存储空间（调试用）
func (cr *ContractRegistry) GetStorage(name string) map[string]string {
	return cr.storages[name]
}

// ============================================================================
// 主函数：演示智能合约
// ============================================================================

func main() {
	fmt.Println("📚 第七课：智能合约")
	fmt.Println("=" + strings.Repeat("=", 49))

	registry := NewContractRegistry()

	// -------------------------------------------------------------------------
	// 演示1：计数器合约
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示1：计数器合约")
	fmt.Println("----------------------------------------")

	counter := NewCounterContract()
	registry.Deploy("Counter", counter)

	// 调用 increment 方法
	result := registry.Execute("Counter", "Alice", "increment", nil)
	printResult("increment()", result)

	result = registry.Execute("Counter", "Alice", "increment", nil)
	printResult("increment()", result)

	result = registry.Execute("Counter", "Alice", "increment", nil)
	printResult("increment()", result)

	// 获取当前值
	result = registry.Execute("Counter", "Alice", "get", nil)
	printResult("get()", result)

	// -------------------------------------------------------------------------
	// 演示2：代币合约
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示2：代币合约")
	fmt.Println("----------------------------------------")

	// 部署代币合约
	token := NewTokenContract("MyToken", "MTK", "Alice")
	registry.Deploy("MyToken", token)

	// Alice 铸造代币给自己
	result = registry.Execute("MyToken", "Alice", "mint", map[string]interface{}{
		"address": "Alice",
		"amount":  "1000",
	})
	printResult("mint(Alice, 1000)", result)

	// Alice 铸造给 Bob
	result = registry.Execute("MyToken", "Alice", "mint", map[string]interface{}{
		"address": "Bob",
		"amount":  "500",
	})
	printResult("mint(Bob, 500)", result)

	// 查询余额
	result = registry.Execute("MyToken", "Alice", "balanceOf", map[string]interface{}{
		"address": "Alice",
	})
	printResult("balanceOf(Alice)", result)

	result = registry.Execute("MyToken", "Alice", "totalSupply", nil)
	printResult("totalSupply()", result)

	// Alice 转账给 Bob
	result = registry.Execute("MyToken", "Alice", "transfer", map[string]interface{}{
		"from":   "Alice",
		"to":     "Bob",
		"amount": "200",
	})
	printResult("transfer(Alice->Bob, 200)", result)

	// 转账后查询余额
	result = registry.Execute("MyToken", "Alice", "balanceOf", map[string]interface{}{
		"address": "Alice",
	})
	printResult("balanceOf(Alice) after transfer", result)

	result = registry.Execute("MyToken", "Alice", "balanceOf", map[string]interface{}{
		"address": "Bob",
	})
	printResult("balanceOf(Bob) after transfer", result)

	// -------------------------------------------------------------------------
	// 演示3：投票合约
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示3：投票合约")
	fmt.Println("----------------------------------------")

	voting := NewVotingContract()
	registry.Deploy("Voting", voting)

	// 添加候选人
	candidates := []string{"Alice", "Bob", "Charlie"}
	for _, candidate := range candidates {
		result = registry.Execute("Voting", "admin", "addCandidate", map[string]interface{}{
			"name": candidate,
		})
	}

	// 用户投票
	voters := []struct {
		voter     string
		candidate string
	}{
		{"user1", "Alice"},
		{"user2", "Alice"},
		{"user3", "Bob"},
		{"user4", "Charlie"},
		{"user5", "Alice"},
		{"user6", "Bob"},
	}

	for _, vote := range voters {
		result = registry.Execute("Voting", vote.voter, "vote", map[string]interface{}{
			"candidate": vote.candidate,
		})
	}

	// 查看结果
	result = registry.Execute("Voting", "admin", "getResults", nil)
	printResult("getResults()", result)

	// 检查是否已投票
	result = registry.Execute("Voting", "admin", "hasVoted", map[string]interface{}{
		"address": "user1",
	})
	printResult("hasVoted(user1)", result)

	// -------------------------------------------------------------------------
	// 演示4：查看合约存储
	// -------------------------------------------------------------------------
	fmt.Println("\n🔹 演示4：合约存储查看")
	fmt.Println("----------------------------------------")

	fmt.Println("\nMyToken 合约存储:")
	storage := registry.GetStorage("MyToken")
	for key, value := range storage {
		if strings.HasPrefix(key, "balance_") {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println("\nVoting 合约存储:")
	storage = registry.GetStorage("Voting")
	for key, value := range storage {
		if strings.HasPrefix(key, "candidate_") {
			name := strings.TrimPrefix(key, "candidate_")
			fmt.Printf("  %s: %s 票\n", name, value)
		}
	}

	// ============================================================================
	// 本课小结
	// ============================================================================
	// ✅ 我们学习了：
	// 1. 智能合约是运行在区块链上的程序
	// 2. 合约有状态（Storage）和逻辑（方法）
	// 3. 实现了计数器、代币、投票三个示例合约
	// 4. 合约注册表用于管理所有已部署的合约
	//
	// ❓ 思考题：
	// 1. 真实的智能合约为什么不能直接修改（增删改函数）？
	// 2. 如何防止合约中的漏洞（如重入攻击）？
	// 3. 智能合约的 gas 费用是什么？
	//
	// 🔜 下一课：P2P 网络 - 学习区块链节点如何通信
	// ============================================================================
	fmt.Println("\n✅ 第七课完成！下一课将学习 P2P 网络。")
}

// ============================================================================
// 辅助函数
// ============================================================================

func printResult(method string, result interface{}) {
	data, _ := json.MarshalIndent(result, "   ", "  ")
	fmt.Printf("   %s:\n", method)
	fmt.Printf("   %s\n", string(data))
}