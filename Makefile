.PHONY: test bins clean
PROJECT_ROOT = github.com/samarabbas/cadence-samples

export PATH := $(GOPATH)/bin:$(PATH)

# default target
default: test

PROGS = helloworld \
	branch \
	childworkflow \
	choice \
	dynamic \
	greetings \
	pickfirst \
	retryactivity \
	splitmerge \
	timer \
	localactivity \
	query \
	cron \
	dsl \
	fileprocessing \
	dummy \
	expense \
	recovery \
	cancelactivity \
TEST_ARG ?= -race -v -timeout 5m
BUILD := ./build
SAMPLES_DIR=./cmd/samples

export PATH := $(GOPATH)/bin:$(PATH)

# Automatically gather all srcs
ALL_SRC := $(shell find ./cmd/samples/common -name "*.go")

# all directories with *_test.go files in them
TEST_DIRS=./cmd/samples/cron \
	./cmd/samples/dsl \
	./cmd/samples/expense \
	./cmd/samples/fileprocessing \
	./cmd/samples/recipes/branch \
	./cmd/samples/recipes/choice \
	./cmd/samples/recipes/greetings \
	./cmd/samples/recipes/helloworld \
	./cmd/samples/recipes/cancel \
	./cmd/samples/recipes/pickfirst \
	./cmd/samples/recipes/retryactivity \
	./cmd/samples/recipes/splitmerge \
	./cmd/samples/recipes/timer \
	./cmd/samples/recipes/localactivity \
	./cmd/samples/recipes/query \
	./cmd/samples/recovery \

dep-ensured:
	dep ensure

cancelactivity: dep-ensured $(ALL_SRC)
	go build -i -o bin/cancelactivity cmd/samples/recipes/cancelactivity/*.go

helloworld: dep-ensured $(ALL_SRC)
	go build -i -o bin/helloworld cmd/samples/recipes/helloworld/*.go

branch: dep-ensured $(ALL_SRC)
	go build -i -o bin/branch cmd/samples/recipes/branch/*.go

childworkflow: dep-ensured $(ALL_SRC)
	go build -i -o bin/childworkflow cmd/samples/recipes/childworkflow/*.go

choice: dep-ensured $(ALL_SRC)
	go build -i -o bin/choice cmd/samples/recipes/choice/*.go

dynamic: dep-ensured $(ALL_SRC)
	go build -i -o bin/dynamic cmd/samples/recipes/dynamic/*.go

greetings: dep-ensured $(ALL_SRC)
	go build -i -o bin/greetings cmd/samples/recipes/greetings/*.go

pickfirst: dep-ensured $(ALL_SRC)
	go build -i -o bin/pickfirst cmd/samples/recipes/pickfirst/*.go

retryactivity: dep-ensured $(ALL_SRC)
	go build -i -o bin/retryactivity cmd/samples/recipes/retryactivity/*.go

splitmerge: dep-ensured $(ALL_SRC)
	go build -i -o bin/splitmerge cmd/samples/recipes/splitmerge/*.go

timer: dep-ensured $(ALL_SRC)
	go build -i -o bin/timer cmd/samples/recipes/timer/*.go

localactivity: dep-ensured $(ALL_SRC)
	go build -i -o bin/localactivity cmd/samples/recipes/localactivity/*.go

query: dep-ensured $(ALL_SRC)
	go build -i -o bin/query cmd/samples/recipes/query/*.go

cron: dep-ensured $(ALL_SRC)
	go build -i -o bin/cron cmd/samples/cron/*.go

dsl: dep-ensured $(ALL_SRC)
	go build -i -o bin/dsl cmd/samples/dsl/*.go

fileprocessing: dep-ensured $(ALL_SRC)
	go build -i -o bin/fileprocessing cmd/samples/fileprocessing/*.go

dummy: dep-ensured $(ALL_SRC)
	go build -i -o bin/dummy cmd/samples/expense/server/*.go

expense: dep-ensured $(ALL_SRC)
	go build -i -o bin/expense cmd/samples/expense/*.go

recovery: dep-ensured $(ALL_SRC)
	go build -i -o bin/recovery cmd/samples/recovery/*.go

bins: helloworld \
	branch \
	childworkflow \
	choice \
	dynamic \
	greetings \
	pickfirst \
	retryactivity \
	splitmerge \
	timer \
	cron \
	dsl \
	fileprocessing \
	dummy \
	expense \
	localactivity \
	query \
	recovery \

test: bins
	@rm -f test
	@rm -f test.log
	@echo $(TEST_DIRS)
	@for dir in $(TEST_DIRS); do \
		go test -coverprofile=$@ "$$dir" | tee -a test.log; \
	done;

clean:
	rm -rf bin
	rm -Rf $(BUILD)
