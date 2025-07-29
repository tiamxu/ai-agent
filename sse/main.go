package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
)

type ChatRequest struct {
	Question string `json:"question"`
	Stream   bool   `json:"stream"`
}

func main() {
	// 创建 Hertz 客户端
	c, err := client.NewClient()
	if err != nil {
		fmt.Printf("创建客户端错误: %v\n", err)
		return
	}

	// 准备请求体
	reqBody := ChatRequest{
		Question: "讲一个笑话",
		Stream:   true,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("JSON 编码错误: %v\n", err)
		return
	}

	// 创建请求和响应对象
	req, resp := protocol.AcquireRequest(), protocol.AcquireResponse()
	defer protocol.ReleaseRequest(req)
	defer protocol.ReleaseResponse(resp)

	// 设置请求
	req.SetRequestURI("http://localhost:8800/api/chat")
	req.SetMethod("POST")
	req.SetBody(jsonBody)
	req.Header.Set("Content-Type", "application/json")
	sse.AddAcceptMIME(req) // 添加 SSE 支持的 MIME 类型

	// 发送请求
	fmt.Println("连接到 SSE 服务器...")
	if err := c.Do(context.Background(), req, resp); err != nil {
		fmt.Printf("请求错误: %v\n", err)
		return
	}

	// 检查响应状态码
	if resp.StatusCode() != 200 {
		fmt.Printf("非预期状态码: %d\n", resp.StatusCode())
		return
	}

	// 创建 SSE 读取器
	r, err := sse.NewReader(resp)
	if err != nil {
		fmt.Printf("创建 SSE 读取器错误: %v\n", err)
		return
	}
	defer r.Close()

	fmt.Println("已连接到 SSE 流. 接收事件中...")

	// 设置信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理中断信号
	go func() {
		<-sigCh
		fmt.Println("\n收到中断信号. 退出中...")
		cancel()
	}()

	// 处理 SSE 事件流
	err = r.ForEach(ctx, func(e *sse.Event) error {
		fmt.Printf("收到事件:\n")
		fmt.Printf("  ID: %s\n", e.ID)
		fmt.Printf("  类型: %s\n", e.Type)
		fmt.Printf("  数据: %s\n\n", string(e.Data))
		return nil
	})

	if err != nil {
		fmt.Printf("处理事件错误: %v\n", err)
	} else {
		fmt.Println("所有事件处理完成")
	}
	time.Sleep(5 * time.Second)

}
