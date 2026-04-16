## On-demand pattern

No workflow is pre-started. The caller creates and controls workflow instances through Nexus
operations. `NexusRemoteGreetingService` adds a `runFromRemote` operation that starts a new
`GreetingWorkflow`, and every other operation includes a user ID so the handler knows which
instance to target.

The caller workflow:
1. Starts two remote `GreetingWorkflow` instances via `runFromRemote` (backed by `WorkflowRunOperation`)
2. Queries each for supported languages
3. Changes the language on each (French and Spanish)
4. Confirms the changes via queries
5. Approves both workflows
6. Waits for each to complete and returns their results

### Running

Start a Temporal server:

```bash
temporal server start-dev
```

Create the namespaces and Nexus endpoint:

```bash
temporal operator namespace create --namespace my-target-namespace
temporal operator namespace create --namespace my-caller-namespace

temporal operator nexus endpoint create \
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

In a third terminal, start the caller workflow:

```bash
go run ./nexus-messaging/ondemandpattern/caller/starter/main.go
```

Expected output:

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
