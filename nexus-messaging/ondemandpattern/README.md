## On-demand pattern

No Workflow is pre-started. The caller creates and controls Workflow instances through Nexus
operations. `NexusRemoteGreetingService` adds a `runFromRemote` operation that starts a new
`GreetingWorkflow`, and every other operation includes a User ID so the handler knows which
instance to target.

The caller :
1. Attaches approval context for the first user via `attachApprovalContext`, before anything has
   started that user's Workflow
2. Starts two remote `GreetingWorkflow` instances via `runFromRemote` (backed by a Workflow started
   through `temporalnexus.StartUntypedWorkflow`)
3. Attaches approval context for the second user, whose Workflow now already exists
4. Queries each for supported languages
5. Changes the language on each (French and Spanish)
6. Confirms the changes via queries
7. Approves both Workflows
8. Waits for each to complete and returns their results

### Running

This sample requires a Temporal dev server build that supports Workflow Update callbacks. Download the compatible
binary from the [Temporal CLI pre-release instructions](https://docs.temporal.io/standalone-nexus-operation#temporal-cli-support).

Start the Temporal dev server with the required namespaces pre-created and Workflow Update callbacks enabled:

```bash
./temporal server start-dev \
  --dynamic-config-value history.enableUpdateCallbacks=true \
  --dynamic-config-value history.enableCHASMSignalBacklinks=true \
  --dynamic-config-value history.enableSignalWithStartFromWorkflow=true \
  --namespace my-target-namespace \
  --namespace my-caller-namespace
```

Create the Nexus endpoint:

```bash
./temporal operator nexus endpoint create \
  --name my-nexus-endpoint-name \
  --target-namespace my-target-namespace \
  --target-task-queue my-handler-task-queue
```

In one terminal, start the handler worker:

```bash
go run ./nexus-messaging/ondemandpattern/handler/worker/main.go
```

In a second terminal, start the caller worker:

```bash
go run ./nexus-messaging/ondemandpattern/caller/worker/main.go
```

In a third terminal, run the following command to start the example:

```bash
go run ./nexus-messaging/ondemandpattern/caller/starter/main.go
```

Expected output:

```
[1] attached approval context before the workflow existed: nexus-messaging-greeting-one
[2] started remote workflow one: nexus-messaging-greeting-one
[3] started remote workflow two: nexus-messaging-greeting-two
[4] attached approval context to the running workflow: nexus-messaging-greeting-two
[5] getLanguages (one) returned 2 languages
[6] getLanguages (two) with unsupported returned 7 languages
[7] setLanguage(French) on one returned previous: English
[8] setLanguage(Spanish) on two returned previous: English
[9] getLanguage (one) = French
[10] getLanguage (two) = Spanish
[11] approved workflow one
[12] approved workflow two
[13] remote workflow one result: Bonjour, monde (approved by CallerRemoteWorkflow)
[14] remote workflow two result: Hola, mundo (approved by CallerRemoteWorkflow)
```
