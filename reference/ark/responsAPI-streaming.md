# demo for streaming response from arkruntime service

```
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

func main() {
	client := arkruntime.NewClientWithApiKey(
		// Get API Key：https://console.volcengine.com/ark/region:ark+cn-beijing/apikey
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
	ctx := context.Background()

	resp, err := client.CreateResponsesStream(ctx, &responses.ResponsesRequest{
		Model:    "doubao-seed-1-6-251015",
		Input:    &responses.ResponsesInput{Union: &responses.ResponsesInput_StringValue{StringValue: "常见的十字花科植物有哪些？"}},
		Thinking: &responses.ResponsesThinking{Type: responses.ThinkingType_enabled.Enum()},
	})
	if err != nil {
		fmt.Printf("stream error: %v", err)
		return
	}
	for {
		event, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("stream error: %v", err)
			return
		}
		handleEvent(event)
	}
}
func handleEvent(event *responses.Event) {
	switch event.GetEventType() {
	case responses.EventType_response_reasoning_summary_text_delta.String():
		print(event.GetReasoningText().GetDelta())
	case responses.EventType_response_reasoning_summary_text_done.String(): // aggregated reasoning text
		fmt.Printf("\nAggregated reasoning text: %s\n", event.GetReasoningText().GetText())
	case responses.EventType_response_output_text_delta.String():
		print(event.GetText().GetDelta())
	case responses.EventType_response_output_text_done.String(): // aggregated output text
		fmt.Printf("\nAggregated output text: %s\n", event.GetTextDone().GetText())
	default:
		return
	}
}
```