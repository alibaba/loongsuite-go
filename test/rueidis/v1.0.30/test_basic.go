package main

import (
	"context"
	"fmt"
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/redis/rueidis"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"
)

func main() {
	// 创建 rueidis 客户端
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{"localhost:" + os.Getenv("REDIS_PORT")}, // Redis 地址
		// Username:  "default",                    // 可选：用户名
		Password: "", // no password set
		// SelectDB:  0,                            // 可选：选择 DB
		DisableCache: true, // 禁用客户端缓存，避免兼容性问题
		DisableRetry: true, // 禁用重试机制
		AlwaysRESP2:  true, // 强制使用RESP2协议，提高兼容性
	})
	if err != nil {
		log.Fatal("Failed to create Redis client:", err.Error())
	}
	defer client.Close()

	ctx := context.Background()

	// 首先测试连接
	fmt.Println("🔍 测试 Redis 连接...")
	pingCmd := client.B().Ping().Build()
	pingResp := client.Do(ctx, pingCmd)
	if pingResp.Error() != nil {
		log.Fatal("Redis 连接失败:", pingResp.Error())
	}
	fmt.Println("✅ Redis 连接成功!")

	// === 1. 基本 SET 和 GET ===
	const key = "greeting"
	if err := client.Do(ctx, client.B().Set().Key(key).Value("Hello from rueidis!").Build()).Error(); err != nil {
		log.Fatal("SET failed:", err)
	}
	fmt.Println("✅ SET greeting = Hello from rueidis!")

	// GET
	getCmd := client.B().Get().Key(key).Build()
	getResp := client.Do(ctx, getCmd)
	if getResp.Error() != nil {
		log.Fatal("GET failed:", getResp.Error())
	}
	r, _ := getResp.ToString()
	fmt.Println("✅ GET greeting =", r)

	// === 2. 使用普通 SET 存储 JSON 字符串 ===
	type User struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"`
	}

	user := User{Name: "Bob", Age: 25, Email: "bob@example.com"}
	jsonKey := "user:1001"

	// 使用普通的 SET 命令存储 JSON 字符串
	userJson := fmt.Sprintf(`{"name":"%s","age":%d,"email":"%s"}`, user.Name, user.Age, user.Email)
	if err := client.Do(ctx, client.B().Set().Key(jsonKey).Value(userJson).Build()).Error(); err != nil {
		fmt.Println("⚠️  SET JSON 失败:", err)
	} else {
		fmt.Println("✅ 使用 SET 存储 JSON:", jsonKey, "=", userJson)

		// 获取并显示结果
		getUserCmd := client.B().Get().Key(jsonKey).Build()
		if getUserResp := client.Do(ctx, getUserCmd); getUserResp.Error() != nil {
			fmt.Println("⚠️  GET JSON 失败:", getUserResp.Error())
		} else {
			userJsonResult, _ := getUserResp.ToString()
			fmt.Println("✅ GET JSON:", userJsonResult)
		}
	}

	// === 3. Pipeline：批量执行多个命令 ===
	fmt.Println("\n🔧 执行批量命令...")
	pipeCmds := []rueidis.Completed{
		client.B().Get().Key("greeting").Build(),
		client.B().Incr().Key("counter").Build(),
		client.B().Exists().Key("greeting").Build(), // 修复：只检查一个键
	}

	responses := client.DoMulti(ctx, pipeCmds...)
	for i, resp := range responses {
		if resp.Error() != nil {
			fmt.Printf("🔧 Multi command %d error: %v\n", i, resp.Error())
		} else {
			// 根据命令类型处理不同的返回值
			switch i {
			case 0: // GET 命令
				result, _ := resp.ToString()
				fmt.Printf("🔧 Multi command %d (GET) result: %s\n", i, result)
			case 1: // INCR 命令
				result, _ := resp.AsInt64()
				fmt.Printf("🔧 Multi command %d (INCR) result: %d\n", i, result)
			case 2: // EXISTS 命令
				result, _ := resp.AsInt64()
				fmt.Printf("🔧 Multi command %d (EXISTS) result: %d\n", i, result)
			default:
				result, _ := resp.ToString()
				fmt.Printf("🔧 Multi command %d result: %s\n", i, result)
			}
		}
	}

	// === 4. 基本的键操作测试 ===
	fmt.Println("\n🔑 测试基本键操作...")

	// 设置过期时间
	if err := client.Do(ctx, client.B().Set().Key("temp_key").Value("temp_value").Ex(60*time.Second).Build()).Error(); err != nil {
		fmt.Println("⚠️  SET with EX failed:", err)
	} else {
		fmt.Println("✅ SET with expiration: temp_key")

		// 检查TTL
		ttlCmd := client.B().Ttl().Key("temp_key").Build()
		if ttlResp := client.Do(ctx, ttlCmd); ttlResp.Error() != nil {
			fmt.Println("⚠️  TTL failed:", ttlResp.Error())
		} else {
			ttl, _ := ttlResp.AsInt64()
			fmt.Printf("✅ TTL temp_key: %d seconds\n", ttl)
		}
	}

	// 发布消息测试（不依赖订阅）
	fmt.Println("\n📢 测试发布消息...")
	if err := client.Do(ctx, client.B().Publish().Channel("news").Message("Hello subscribers!").Build()).Error(); err != nil {
		fmt.Println("⚠️  PUBLISH failed:", err)
	} else {
		fmt.Println("✅ PUBLISH message to channel 'news'")
	}

	// 显示一些 Redis 信息
	fmt.Println("\n📊 获取 Redis 信息...")
	infoCmd := client.B().Info().Section("server").Build()
	if infoResp := client.Do(ctx, infoCmd); infoResp.Error() != nil {
		fmt.Println("⚠️  INFO failed:", infoResp.Error())
	} else {
		info, _ := infoResp.ToString()
		fmt.Println("✅ Redis Server Info:")
		// 只显示前几行
		lines := fmt.Sprintf("%.200s...", info)
		fmt.Println(lines)
	}

	fmt.Println("\n🎉 所有测试完成!")
	time.Sleep(2 * time.Second)
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		traceNum := len(stubs)
		verifier.Assert(traceNum == 10, "Expected 10 trace num, got %d", traceNum)
		pingSpan := stubs[0][0]
		verifier.Assert(pingSpan.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", pingSpan.SpanKind)
		verifier.Assert(pingSpan.Name == "PING", "Except server span name to be ping, got %s", pingSpan.Name)
		setSpan := stubs[1][0]
		verifier.Assert(setSpan.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", setSpan.SpanKind)
		verifier.Assert(setSpan.Name == "SET", "Except server span name to be set, got %s", setSpan.Name)
		getSpan := stubs[2][0]
		verifier.Assert(getSpan.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", getSpan.SpanKind)
		verifier.Assert(getSpan.Name == "GET", "Except server span name to be get, got %s", getSpan.Name)
		setSpan1 := stubs[3][0]
		verifier.Assert(setSpan1.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", setSpan1.SpanKind)
		verifier.Assert(setSpan1.Name == "SET", "Except server span name to be set, got %s", setSpan1.Name)
		getSpan1 := stubs[4][0]
		verifier.Assert(getSpan1.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", getSpan1.SpanKind)
		verifier.Assert(getSpan1.Name == "GET", "Except server span name to be get, got %s", getSpan1.Name)
		mutilSpan := stubs[5][0]
		verifier.Assert(mutilSpan.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", mutilSpan.SpanKind)
		verifier.Assert(mutilSpan.Name == "GET INCR EXISTS", "Except server span name to be GET INCR EXISTS, got %s", mutilSpan.Name)
		setSpan2 := stubs[6][0]
		verifier.Assert(setSpan2.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", setSpan2.SpanKind)
		verifier.Assert(setSpan2.Name == "SET", "Except server span name to be set, got %s", setSpan2.Name)
		pubSpan := stubs[8][0]
		verifier.Assert(pubSpan.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", pubSpan.SpanKind)
		verifier.Assert(pubSpan.Name == "PUBLISH", "Except server span name to be PUBLISH, got %s", pubSpan.Name)
	}, 1)
}
