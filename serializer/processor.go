package serializer

import (
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type (
	EventProcessor struct {
		logger  log.Logger
		eventCh workflow.ReceiveChannel

		state *EventProcessorState
	}

	EventProcessorState struct {
		PreviousEventID int
		Items           map[int]ResourceEvent
	}
)

func newEventProcessor(ch workflow.ReceiveChannel, prevState *EventProcessorState) *EventProcessor {
	prevEventID := 0
	items := make(map[int]ResourceEvent)
	if prevState != nil {
		prevEventID = prevState.PreviousEventID
		if len(prevState.Items) > 0 {
			items = prevState.Items
		}
	}

	return &EventProcessor{
		eventCh: ch,
		state: &EventProcessorState{
			PreviousEventID: prevEventID,
			Items:           items,
		},
	}
}

func (p *EventProcessor) start(rootCtx workflow.Context) workflow.ReceiveChannel {
	p.logger = workflow.GetLogger(rootCtx)
	// doneCh is used to communicate unprocessed events back to the workflow after shutdown
	// Workflow passes these unprocessed events to the new run so they can be processed after missing events comes in
	doneCh := workflow.NewChannel(rootCtx)
	workflow.Go(rootCtx, func(ctx workflow.Context) {
		p.pump(ctx, doneCh)
	})

	return doneCh
}

func (p *EventProcessor) pump(ctx workflow.Context, doneCh workflow.SendChannel) {
	done := false
	for !done {
		var event ResourceEvent
		more := p.eventCh.Receive(ctx, &event)
		if !more {
			// Channel is closed.  Workflow wants to shutdown.
			done = true
			break
		}

		p.state.addEvent(event, p.logger)
		p.processInOrder(ctx)
	}

	doneCh.Send(ctx, p.state)
}

func (p *EventProcessor) processInOrder(ctx workflow.Context) {
	s := p.state
	nextID := s.PreviousEventID + 1
	for current, ok := s.Items[nextID]; ok; current, ok = s.Items[nextID] {
		p.logger.Info("Processing event",
			"EventID", current.EventID)
		err := workflow.ExecuteActivity(ctx, ProcessEvent, current).Get(ctx, nil)
		if err != nil {
			p.logger.Error("Failed to process event",
				"EventID", current.EventID,
				"Error", err)
		}

		delete(s.Items, nextID)
		nextID++
	}

	s.PreviousEventID = nextID - 1
}

func (s *EventProcessorState) addEvent(event ResourceEvent, logger log.Logger) {
	if event.EventID <= s.PreviousEventID {
		logger.Info("Dedupe already processed event",
			"LastEventID", s.PreviousEventID,
			"EventID", event.EventID,
		)

		return
	}

	nextID := s.PreviousEventID + 1
	if event.EventID > nextID {
		logger.Info("Out of order event",
			"LastEventID", s.PreviousEventID,
			"EventID", event.EventID)
	}

	// Store event
	s.Items[event.EventID] = event
}
