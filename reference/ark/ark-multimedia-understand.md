# demo code for ark multimedia understanding

```
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/file"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

func main() {
	client := arkruntime.NewClientWithApiKey(os.Getenv("ARK_API_KEY"))
	ctx := context.Background()

	fmt.Println("----- upload file data -----")
	data, err := os.Open("/Users/doc/demo.pdf")
	if err != nil {
		fmt.Printf("read file error: %v\n", err)
		return
	}
	fileInfo, err := client.UploadFile(ctx, &file.UploadFileRequest{
		File:    data,
		Purpose: file.PurposeUserData,
	})

	if err != nil {
		fmt.Printf("upload file error: %v", err)
		return
	}

	// Wait for the file to finish processing
	for fileInfo.Status == file.StatusProcessing {
		fmt.Println("Waiting for file to be processed...")
		time.Sleep(2 * time.Second)
		fileInfo, err = client.RetrieveFile(ctx, fileInfo.ID) // update file info
		if err != nil {
			fmt.Printf("get file status error: %v", err)
			return
		}
	}
	fmt.Printf("Video processing completed: %s, status: %s\n", fileInfo.ID, fileInfo.Status)
	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_File{
					File: &responses.ContentItemFile{
						Type:   responses.ContentItemType_input_file,
						FileId: volcengine.String(fileInfo.ID),
					},
				},
			},
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: "按段落给出文档中的文字内容，以JSON格式输出，包括段落类型（type）、文字内容（content）信息。",
					},
				},
			},
		},
	}
	createResponsesReq := &responses.ResponsesRequest{
		Model: "doubao-seed-1-6-251015",
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
		Caching: &responses.ResponsesCaching{Type: responses.CacheType_enabled.Enum()},
	}

	resp, err := client.CreateResponsesStream(ctx, createResponsesReq)
	if err != nil {
		fmt.Printf("stream error: %v\n", err)
		return
	}
	var responseId string
	for {
		event, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("stream error: %v\n", err)
			return
		}
		handleEvent(event)
		if responseEvent := event.GetResponse(); responseEvent != nil {
			responseId = responseEvent.GetResponse().GetId()
			fmt.Printf("Response ID: %s", responseId)
		}
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