# Nexus Messaging — Caller Pattern (Entity Pattern)

This example shows the **entity pattern**: the handler worker pre-starts a `GreetingWorkflow` for a user at boot. Nexus operations receive a `userID` and route to the long-running workflow via the `GreetingWorkflow_for_<userID>` workflow ID prefix.

## Operations (NexusGreetingService)

| Operation | Type | Description |
|-----------|------|-------------|
| `getLanguages` | Sync | Queries the entity workflow for supported languages |
| `getLanguage` | Sync | Queries the entity workflow for the current language |
| `setLanguage` | Sync | Sends an update to change the language |
| `approve` | Sync | Sends a signal to approve and complete the workflow |

## Running

For more details on Nexus and how to set up to run this sample, please see the [Nexus Sample](../nexus/README.md).

### 1. Start the handler worker (pre-starts the entity workflow)

```bash
go run ./nexus-messaging/callerpattern/handler/worker/main.go
```

### 2. Start the caller worker

```bash
go run ./nexus-messaging/callerpattern/caller/worker/main.go \
  -namespace my-caller-namespace
```

### 3. Run the caller workflow

```bash
go run ./nexus-messaging/callerpattern/caller/starter/main.go \
  -namespace my-caller-namespace
```

The caller workflow will:
1. Call `getLanguages` to retrieve supported languages
2. Call `setLanguage(Arabic)` to change the language
3. Call `getLanguage` to verify the language is Arabic
4. Call `approve` to signal and complete the entity workflow

### 4. Output

which should result in:
```
[1] getLanguages returned 2 languages
[2] setLanguage(Arabic) returned previous language: English
[3] getLanguage returned: Arabic (confirmed Arabic)
[4] approve sent successfully
```