package main

import (
	"errors"

	"yact/api"
	"yact/config"
	"yact/logic"
)

var errRecordingClientNoBatches = errors.New("the recording LLM client does not support batches")

// Tests substitute this client through the ClientFactory seam, so no network call is made.
type recordingLLMClient struct {
	reply string
	calls int
}

func newRecordingLLMClient(reply string) *recordingLLMClient {
	return &recordingLLMClient{reply: reply}
}

func (c *recordingLLMClient) newClientFactory() api.ClientFactory {
	return func(cfg *config.Config) api.Client {
		return c
	}
}

func (c *recordingLLMClient) GetModelName() string {
	return "recording"
}

func (c *recordingLLMClient) Call(requestContent *logic.RequestContent, systemPrompt string) (api.CallResult, error) {
	c.calls++
	return api.CallResult{Text: c.reply}, nil
}

func (c *recordingLLMClient) SubmitBatch(requestContent *logic.RequestContent, systemPrompt string) (string, error) {
	return "", errRecordingClientNoBatches
}

func (c *recordingLLMClient) GetBatchStatus(batchID string) (api.BatchStatus, error) {
	return api.BatchStatus{}, errRecordingClientNoBatches
}

func (c *recordingLLMClient) GetBatchResult(batchID string) (api.BatchResult, error) {
	return api.BatchResult{}, errRecordingClientNoBatches
}
