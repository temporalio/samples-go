This sample shows how to expose a long-running workflow's queries, updates, and signals as Nexus
operations. There are two self-contained examples, each in its own directory:

| | `callerpattern/` | `ondemandpattern/` |
|---|---|---|
| **Pattern** | Signal an existing workflow | Create and run workflows on demand, and send signals to them |
| **Who creates the workflow?** | The handler worker starts it on boot | The caller starts it via a Nexus operation |
| **Who knows the workflow ID?** | Only the handler | The caller chooses and passes it in every operation |
| **Nexus service** | `NexusGreetingService` | `NexusRemoteGreetingService` |

Each directory is fully self-contained for clarity. The
`GreetingWorkflow`, `GreetingActivity`, and `Language` type are pretty much the same between the two -- only the
Nexus service definition and its handler implementation differ. This highlights that the same workflow can be
exposed through Nexus in different ways depending on whether the caller needs lifecycle control.

See each directory's README for running instructions.
