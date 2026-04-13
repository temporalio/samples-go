# Nexus Messaging Sample

This sample demonstrates how to expose a long-running workflow's queries, updates, and signals as Nexus operations. It contains two self-contained sub-examples.

## Shared Concepts

Both patterns share:

- A `GreetingWorkflow` that is a long-running entity workflow supporting:
  - Query `getLanguages` — returns supported languages
  - Query `getLanguage` — returns the current language
  - Update `setLanguage` — changes the language (calls an activity for unknown languages)
  - Signal `approve` — approves the workflow, causing it to complete
- A `GreetingActivity` that returns greetings for all 7 languages
- **Languages**: Arabic, Chinese, English, French, Hindi, Portuguese, Spanish
- **Initial greetings map**: `{Chinese: "你好，世界", English: "Hello, world"}`
- **Namespaces**: `my-target-namespace`, `my-caller-namespace`
- **Endpoint**: `my-nexus-endpoint-name`
- **Handler task queue**: `my-handler-task-queue`
- my-handler-task-queue

## Sub-examples

### [callerpattern](./callerpattern/README.md) — Entity Pattern

The handler worker pre-starts a `GreetingWorkflow` for each user at boot. Nexus operations receive a `userID` and route to the corresponding workflow via a `GreetingWorkflow_for_<userID>` prefix.

### [ondemandpattern](./ondemandpattern/README.md) — On-Demand Pattern

No workflow is pre-started. The caller creates workflows on demand via the `runFromRemote` operation (`WorkflowRunOperation`). Other operations take a `workflowID` directly.

## Prerequisites

- A running Temporal server with two namespaces and a Nexus endpoint configured:

```bash
temporal operator namespace create my-target-namespace
temporal operator namespace create my-caller-namespace
temporal operator nexus endpoint create \
  --name my-nexus-endpoint-name \
  --target-namespace my-target-namespace \
  --target-task-queue my-handler-task-queue
```
