// Package streams contains the workflows, activities, and shared types for the
// workflow streams sample. A workflow stream is a durable publish/subscribe log
// hosted inside a Temporal workflow: external code publishes to named topics via
// signals, subscribers long-poll for new items via updates, and a query exposes
// the current offset. See go.temporal.io/sdk/contrib/workflowstreams.
package streams

import (
	"time"

	"go.temporal.io/sdk/contrib/workflowstreams"
)

// Task queues. The LLM scenario runs on its own queue so its OpenAI dependency
// stays isolated from the other workflows.
const (
	TaskQueue    = "workflow-streams"
	LLMTaskQueue = "workflow-streams-llm"
)

// Topic names used across the scenarios.
const (
	TopicStatus   = "status"
	TopicProgress = "progress"
	TopicNews     = "news"
	TopicTick     = "tick"
	TopicDelta    = "delta"
	TopicComplete = "complete"
	TopicRetry    = "retry"
)

// CloseSignal tells the hub workflow (scenario 3) to stop hosting its stream.
const CloseSignal = "close"

// Each workflow input carries an optional *workflowstreams.WorkflowStreamState
// so the stream can survive continue-as-new: thread the prior run's state back
// in and pass it to NewWorkflowStream. It is nil on a fresh start.

// OrderInput is the input to OrderWorkflow (scenario 1).
type OrderInput struct {
	OrderID     string                               `json:"orderId"`
	StreamState *workflowstreams.WorkflowStreamState `json:"streamState,omitempty"`
}

// PipelineInput is the input to PipelineWorkflow (scenario 2).
type PipelineInput struct {
	PipelineID  string                               `json:"pipelineId"`
	StreamState *workflowstreams.WorkflowStreamState `json:"streamState,omitempty"`
}

// HubInput is the input to HubWorkflow (scenario 3).
type HubInput struct {
	HubID       string                               `json:"hubId"`
	StreamState *workflowstreams.WorkflowStreamState `json:"streamState,omitempty"`
}

// TickerInput is the input to TickerWorkflow (scenario 4). Zero-valued fields
// fall back to the defaults applied in the workflow.
type TickerInput struct {
	Count         int                                  `json:"count"`
	KeepLast      int                                  `json:"keepLast"`
	TruncateEvery int                                  `json:"truncateEvery"`
	Interval      time.Duration                        `json:"interval"`
	StreamState   *workflowstreams.WorkflowStreamState `json:"streamState,omitempty"`
}

// LLMInput is the input to LLMWorkflow (scenario 5).
type LLMInput struct {
	Prompt      string                               `json:"prompt"`
	Model       string                               `json:"model"`
	StreamState *workflowstreams.WorkflowStreamState `json:"streamState,omitempty"`
}

// Event types published to the stream. They are JSON-encoded by the default data
// converter on the way in and decoded by subscribers on the way out.

// StatusEvent reports an order's lifecycle stage on TopicStatus.
type StatusEvent struct {
	Kind    string `json:"kind"`
	OrderID string `json:"orderId"`
}

// ProgressEvent reports fine-grained progress on TopicProgress.
type ProgressEvent struct {
	Message string `json:"message"`
}

// StageEvent reports a pipeline stage on TopicStatus.
type StageEvent struct {
	Stage string `json:"stage"`
}

// NewsEvent is published by an external publisher on TopicNews.
type NewsEvent struct {
	Headline string `json:"headline"`
}

// TickEvent is published by the ticker on TopicTick.
type TickEvent struct {
	N int `json:"n"`
}

// TextDelta is a single streamed token chunk on TopicDelta.
type TextDelta struct {
	Text string `json:"text"`
}

// TextComplete is the final accumulated completion on TopicComplete.
type TextComplete struct {
	FullText string `json:"fullText"`
}

// RetryEvent signals that the streaming activity is on a retry attempt, so
// subscribers can reset any partially rendered output.
type RetryEvent struct {
	Attempt int `json:"attempt"`
}
