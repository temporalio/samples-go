### Samples for Cadence
This samples folder consist of several samples for cadence. Each subfolder is a separated sample case.

### Prerequisite
Run Cadence Server

See instructions for running the Cadence Server: https://github.com/uber/cadence/blob/master/README.md

### Steps to run samples
Make sure Cadence Server is running.
#### There are several samples for you to get started, they are under cmd/samples folder. Each subfolder defines a sample case of workflow. Run "make" to build. 
#### Run "./helloworld -m worker" to start workers for helloworld workflow.
#### Run "./helloworld -m trigger" to submit a start request for helloworld workflow.

See more documentation at: https://github.com/uber-go/cadence-client

### Commands to run other samples:

#### cron
"./cron -m worker"
"./cron -m trigger -i 3 -c 5" start workflow with interval of 3s and schedule 5 times for the cron job.

#### dsl
"./dsl -m worker"
"./dsl -m trigger -dslConfig cmd/samples/dsl/workflow1.yaml" run workflow1.yaml
"./dsl -m trigger -dslConfig cmd/samples/dsl/workflow2.yaml" run workflow2.yaml

#### expense
  see more details in README.md under expense folder

#### fileprocessing
"./fileprocessing -m worker"
"./fileprocessing -m trigger"

#### recipes/branch
"./branch -m worker"
"./branch -m trigger -c branch" this will run the branch workflow
"./branch -m trigger -c parallel" this will run the parallel branch workflow

#### recipes/choice
"./choice -m worker"
"./choice -m trigger -c single" this will run the single choice workflow
"./choice -m trigger -c multi" this will run the multi choice workflow

#### greetings
"./greetings -m worker"
"./greetings -m trigger"

#### helloworld
"./helloworld -m worker"
"./helloworld -m trigger"

#### pickfirst
"./pickfirst -m worker"
"./pickfirst -m trigger"

#### retryactivity
"./retryactivity -m worker"
"./retryactivity -m trigger"

#### splitmerge
"./splitmerge -m worker"
"./splitmerge -m trigger"

#### timer
"./timer -m worker"
"./timer -m trigger"

#### childworkflow
"./childworkflow -m worker"
"./childworkflow -m trigger"

#### dynamic
"./dynamic -m worker"
"./dynamic -m trigger"
