############################# Main targets #############################
# Run all checks, build, and test.
install: clean lint workflowcheck bins test
########################################################################

##### Variables ######
UNIT_TEST_DIRS := $(sort $(dir $(shell find . -name "*_test.go")))
MAIN_FILES := $(shell find . -name "main.go")
TEST_TIMEOUT := 20s
COLOR := "\e[1;36m%s\e[0m\n"

define NEWLINE


endef

##### Targets ######
bins:
	@printf $(COLOR) "Build samples..."
	$(foreach MAIN_FILE,$(MAIN_FILES), go build -o bin/$(shell dirname "$(MAIN_FILE)") $(shell dirname "$(MAIN_FILE)")$(NEWLINE))

test:
	@printf $(COLOR) "Run unit tests..."
	@rm -f test.log
	$(foreach UNIT_TEST_DIR,$(UNIT_TEST_DIRS),\
		@go test -timeout $(TEST_TIMEOUT) -race $(UNIT_TEST_DIR) | tee -a test.log \
	$(NEWLINE))
	@! grep -q "^--- FAIL" test.log

lint:
	@printf $(COLOR) "Run checks..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	@golangci-lint run --disable-all -E errcheck -E staticcheck

workflowcheck:
	@printf $(COLOR) "Run workflow check..."
	@go install go.temporal.io/sdk/contrib/tools/workflowcheck
	@workflowcheck -config workflowcheck.config.yaml -show-pos ./...

update-sdk:
	go get -u go.temporal.io/api@master
	go get -u go.temporal.io/sdk@master
	go mod tidy

clean:
	rm -rf bin

ci-build: lint workflowcheck bins test
