package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
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
		Desc: "阿里云DNS解析记录操作工具，可以查询、创建、更新、删除DNS记录",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"action": {
				Desc:     "操作类型: query, create, update, delete, enable, disable",
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
	case "create":
		return d.createRecord(ctx, params.Domain, params.RR, params.Type, params.Value, params.TTL)
	default:
		return "", fmt.Errorf("不支持的操作类型")
	}
}

func (d *DNSTool) queryRecord(ctx context.Context, domain, rr string) (string, error) {
	log.Printf("查询DNS记录: domain=%s, rr=%s", domain, rr)
	return "", nil
}

func (d *DNSTool) createRecord(ctx context.Context, domain, rr, recordType, value string, ttl int) (string, error) {
	log.Printf("创建DNS记录: domain=%s, rr=%s, type=%s, value=%s, ttl=%d", domain, rr, recordType, value, ttl)
	return "", nil
}
