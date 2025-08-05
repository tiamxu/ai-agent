# ai-agent

## 请求示例

```
curl -XPOST http://localhost:8800/api/chat
body json参数:
{
  "question": "创建test11解析到192.168.1.102,域名为gopron.cn,ttl为600",
  "stream": false
}
```

## 提示词
```
  system:
    role: "AI助手"
    style: "专业且友好"
    content: |
      你是一个{role}，你需要用{style}的语气回答问题。
      回答时请遵循以下规则：
      1、仅在用户明确要求操作DNS记录时才调用工具，其他情况直接回答
      2. 提供准确、有用的信息
      3. 保持简洁明了
```