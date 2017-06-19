## Cadence
Cadence is a distributed, scalable, durable, and highly available orchestration engine we developed at Uber Engineering to execute asynchronous long-running business logic in a scalable and resilient way.

These are some samples to demostrate various capabilities of Cadence client and server.  You can learn more more about cadence at:
* Cadence: https://github.com/uber/cadence
* Cadence Client: https://github.com/uber-go/cadence-client

## Cadence Samples
This samples folder consist of several samples for cadence. Each subfolder is a separated sample case.

### Prerequisite
Run Cadence Server

See instructions for running the Cadence Server: https://github.com/uber/cadence/blob/master/README.md

### Steps to run samples
Make sure Cadence Server is running.
#### There are several samples for you to get started, they are under cmd/samples folder. Each subfolder defines a sample case of workflow. Run "make" to build. 
#### Run "./bin/helloworld -m worker" to start workers for helloworld workflow.
#### Run "./bin/helloworld -m trigger" to submit a start request for helloworld workflow.

See more documentation at: https://github.com/uber-go/cadence-client

### Commands to run other samples:

#### cron
"./bin/cron -m worker"
"./bin/cron -m trigger -i 3 -c 5" start workflow with interval of 3s and schedule 5 times for the cron job.

#### dsl
"./bin/dsl -m worker"
"./bin/dsl -m trigger -dslConfig cmd/samples/dsl/workflow1.yaml" run workflow1.yaml
"./bin/dsl -m trigger -dslConfig cmd/samples/dsl/workflow2.yaml" run workflow2.yaml

#### expense
  see more details in README.md under expense folder

#### fileprocessing
"./bin/fileprocessing -m worker"
"./bin/fileprocessing -m trigger"

#### recipes/branch
"./bin/branch -m worker"
"./bin/branch -m trigger -c branch" this will run the branch workflow
"./bin/branch -m trigger -c parallel" this will run the parallel branch workflow

#### recipes/choice
"./bin/choice -m worker"
"./bin/choice -m trigger -c single" this will run the single choice workflow
"./bin/choice -m trigger -c multi" this will run the multi choice workflow

#### greetings
"./bin/greetings -m worker"
"./bin/greetings -m trigger"

#### helloworld
"./bin/helloworld -m worker"
"./bin/helloworld -m trigger"

#### pickfirst
"./bin/pickfirst -m worker"
"./bin/pickfirst -m trigger"

#### retryactivity
"./bin/retryactivity -m worker"
"./bin/retryactivity -m trigger"

#### splitmerge
"./bin/splitmerge -m worker"
"./bin/splitmerge -m trigger"

#### timer
"./bin/timer -m worker"
"./bin/timer -m trigger"

#### childworkflow
"./bin/childworkflow -m worker"
"./bin/childworkflow -m trigger"

#### dynamic
"./bin/dynamic -m worker"
"./bin/dynamic -m trigger"
