package llm

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func (s *LLMService) ReportStream(sr *schema.StreamReader[*schema.Message]) error {

	defer sr.Close()

	i := 0
	for {
		message, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("接收流数据失败: %w", err)
		}
		if message == nil || len(message.Content) == 0 {
			continue
		}
		log.Printf("[Message %d] Content: %s\n", i, strings.TrimSpace(message.Content))
		i++
	}
	return nil
}
