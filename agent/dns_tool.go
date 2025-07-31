package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/tiamxu/ai-agent/types"
)

type DNSTool struct {
	BaseURL string
}

func NewDNSTool(baseURL string) *DNSTool {
	return &DNSTool{BaseURL: baseURL}
}
func (d *DNSTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "aliyun_dns_operator",
		Desc: "阿里云DNS解析记录操作工具",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Desc:     "操作类型: query, add, update, delete, enable, disable",
				Type:     schema.String,
				Required: true,
			},
			"domain": {
				Desc:     "域名",
				Type:     schema.String,
				Required: true,
			},
			"rr": {
				Desc: "主机记录",
				Type: schema.String,
			},
			"type": {
				Desc: "记录类型: A, CNAME等",
				Type: schema.String,
			},
			"value": {
				Desc: "记录值",
				Type: schema.String,
			},
			"ttl": {
				Desc: "TTL时间",
				Type: schema.Integer,
			},
			"status": {
				Desc: "状态: enable/disable",
				Type: schema.String,
			},
			"record_id": {
				Desc: "记录ID(更新时需要)",
				Type: schema.String,
			},
		}),
	}, nil
}

func (d *DNSTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params struct {
		Action   string `json:"action"`
		Domain   string `json:"domain"`
		RR       string `json:"rr"`
		Type     string `json:"type"`
		Value    string `json:"value"`
		TTL      int    `json:"ttl"`
		Status   string `json:"status"`
		RecordID string `json:"record_id"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("解析参数失败: %v", err)
	}
	fmt.Printf("params:%v \n", params)
	switch params.Action {
	case "query":
		return d.queryRecord(ctx, params.Domain, params.RR)
	case "add":
		return d.addRecord(ctx, params.Domain, params.RR, params.Type, params.Value, params.TTL)
	default:
		return "", fmt.Errorf("不支持的操作类型")
	}
}

func (d *DNSTool) queryRecord(ctx context.Context, domain, rr string) (string, error) {
	log.Printf("查询DNS记录: domain=%s, rr=%s", domain, rr)
	url := fmt.Sprintf("%s/api/dns/records?domain=%s&rr=%s", d.BaseURL, domain, rr)
	if rr != "" {
		url += fmt.Sprintf("&rr=%s", rr)
	}
	fmt.Printf("url:%v\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("查询DNS记录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	return string(body), nil
}

func (d *DNSTool) createRecord(ctx context.Context, domain, rr, recordType, value string, ttl int) (json.RawMessage, error) {
	return nil, nil
}

func (d *DNSTool) addRecord(ctx context.Context, domain, rr, recordType, value string, ttl int) (string, error) {
	log.Printf("创建DNS记录: domain=%s, rr=%s, type=%s, value=%s, ttl=%d", domain, rr, recordType, value, ttl)

	record := types.DNSRecord{
		Domain: domain,
		RR:     rr,
		Type:   recordType,
		Value:  value,
		TTL:    ttl,
	}

	jsonData, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("编码请求数据失败: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/dns/records", d.BaseURL),
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		return "", fmt.Errorf("添加DNS记录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	return string(body), nil
}
