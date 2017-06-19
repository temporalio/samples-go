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
	cron \
	dsl \
	fileprocessing \
	dummy \
	expense \
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
	./cmd/samples/recipes/pickfirst \
	./cmd/samples/recipes/retryactivity \
	./cmd/samples/recipes/splitmerge \
	./cmd/samples/recipes/timer \

vendor/glide.updated: glide.lock glide.yaml
	glide install
	touch vendor/glide.updated

helloworld: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/helloworld cmd/samples/recipes/helloworld/*.go

branch: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/branch cmd/samples/recipes/branch/*.go

childworkflow: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/childworkflow cmd/samples/recipes/childworkflow/*.go

choice: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/choice cmd/samples/recipes/choice/*.go

dynamic: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/dynamic cmd/samples/recipes/dynamic/*.go

greetings: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/greetings cmd/samples/recipes/greetings/*.go

pickfirst: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/pickfirst cmd/samples/recipes/pickfirst/*.go

retryactivity: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/retryactivity cmd/samples/recipes/retryactivity/*.go

splitmerge: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/splitmerge cmd/samples/recipes/splitmerge/*.go

timer: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/timer cmd/samples/recipes/timer/*.go

cron: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/cron cmd/samples/cron/*.go

dsl: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/dsl cmd/samples/dsl/*.go

fileprocessing: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/fileprocessing cmd/samples/fileprocessing/*.go

dummy: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/dummy cmd/samples/expense/server/*.go

expense: vendor/glide.updated $(ALL_SRC)
	go build -i -o bin/expense cmd/samples/expense/*.go

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