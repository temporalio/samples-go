# Nexus Messaging — On-Demand Pattern

This example shows the **on-demand pattern**: no workflow is pre-started. The caller creates `GreetingWorkflow` instances on demand via the `runFromRemote` operation (`WorkflowRunOperation`). The other operations take a `workflowID` directly to route to a specific workflow instance.

## Operations (NexusRemoteGreetingService)

| Operation | Type | Description |
|-----------|------|-------------|
| `runFromRemote` | Async (WorkflowRunOperation) | Starts a new GreetingWorkflow |
| `getLanguages` | Sync | Queries the workflow for supported languages |
| `getLanguage` | Sync | Queries the workflow for the current language |
| `setLanguage` | Sync | Sends an update to change the language |
| `approve` | Sync | Sends a signal to approve and complete the workflow |

## Running

For more details on Nexus and how to set up to run this sample, please see the [Nexus Sample](../nexus/README.md).

### 1. Start the handler worker

```bash
go run ./nexus-messaging/ondemandpattern/handler/worker/main.go
```

### 2. Start the caller worker

```bash
go run ./nexus-messaging/ondemandpattern/caller/worker/main.go \
  -namespace my-caller-namespace
```

### 3. Run the caller workflow

```bash
go run ./nexus-messaging/ondemandpattern/caller/starter/main.go \
  -namespace my-caller-namespace
```

The caller workflow will:
1. Start two remote `GreetingWorkflow` instances via `runFromRemote`
2. Query languages from both workflows
3. Set language to French on workflow one, Spanish on workflow two
4. Confirm the current language on both
5. Approve both workflows
6. Wait for both workflows to return their greeting results

### 4. Output

which should result in:
```
[1] started remote workflow one: nexus-messaging-greeting-one
[2] started remote workflow two: nexus-messaging-greeting-two
[3] getLanguages (one) returned 2 languages
[4] getLanguages (two) with unsupported returned 7 languages
[5] setLanguage(French) on one returned previous: English
[6] setLanguage(Spanish) on two returned previous: English
[7] getLanguage (one) = French
[8] getLanguage (two) = Spanish
[9] approved workflow one
[10] approved workflow two
[11] remote workflow one result: Bonjour, monde (approved by CallerRemoteWorkflow)
[12] remote workflow two result: Hola, mundo (approved by CallerRemoteWorkflow)
```
