package streams

import (
	"context"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/contrib/workflowstreams"
)

// StreamCompletion (scenario 5) calls OpenAI with streaming enabled and republishes
// each token delta to the workflow stream. The accumulated text is published on
// the complete topic and returned as the activity result. Because the activity
// owns the non-deterministic OpenAI call, the workflow stays deterministic.
//
// maxRetries is set to 0 on the OpenAI client so transient failures surface as
// Temporal activity retries instead. On a retry (attempt > 1) it publishes a
// RetryEvent so subscribers can reset partially rendered output.
func StreamCompletion(ctx context.Context, input LLMInput) (string, error) {
	c, err := workflowstreams.NewClientFromActivity(ctx, workflowstreams.Options{
		BatchInterval: 200 * time.Millisecond,
	})
	if err != nil {
		return "", err
	}
	defer func() { _ = c.Close(ctx) }()

	deltas := c.Topic(TopicDelta)
	complete := c.Topic(TopicComplete)
	retry := c.Topic(TopicRetry)

	if attempt := activity.GetInfo(ctx).Attempt; attempt > 1 {
		retry.Publish(RetryEvent{Attempt: int(attempt)}, true)
	}

	model := input.Model
	if model == "" {
		model = openai.ChatModelGPT4oMini
	}

	client := openai.NewClient(option.WithMaxRetries(0))
	stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:    model,
		Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage(input.Prompt)},
	})
	defer func() { _ = stream.Close() }()

	var full strings.Builder
	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) == 0 {
			continue
		}
		text := chunk.Choices[0].Delta.Content
		if text == "" {
			continue
		}
		deltas.Publish(TextDelta{Text: text}, false)
		full.WriteString(text)
	}
	if err := stream.Err(); err != nil {
		return "", err
	}

	fullText := full.String()
	complete.Publish(TextComplete{FullText: fullText}, true)
	return fullText, nil
}
