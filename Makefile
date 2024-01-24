############################# Main targets #############################
# Run all checks, build, and test.
install: clean staticcheck errcheck workflowcheck bins test
########################################################################

##### Variables ######
UNIT_TEST_DIRS := $(sort $(dir $(shell find . -name "*_test.go")))
MAIN_FILES := $(shell find . -name "main.go")
MOD_FILES_DIR := $(sort $(dir $(shell find . -name "go.mod")))
TEST_TIMEOUT := 20s
COLOR := "\e[1;36m%s\e[0m\n"
BIN_DIR := $(shell pwd)/bin

define NEWLINE


endef

##### Targets ######
bins:
	@printf $(COLOR) "Build samples..."
	$(foreach MAIN_FILE,$(MAIN_FILES), cd $(shell dirname "$(MAIN_FILE)") && go build -o $(BIN_DIR)/$(shell dirname "$(MAIN_FILE)") $(NEWLINE))

test:
	@printf $(COLOR) "Run unit tests..."
	@rm -f test.log
	$(foreach UNIT_TEST_DIR,$(UNIT_TEST_DIRS),\
		@go test -timeout $(TEST_TIMEOUT) -race $(UNIT_TEST_DIR) | tee -a test.log \
	$(NEWLINE))
	@! grep -q "^--- FAIL" test.log

staticcheck:
	@printf $(COLOR) "Run static check..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...

errcheck:
	@printf $(COLOR) "Run error check..."
	@GO111MODULE=off go get -u github.com/kisielk/errcheck
	@errcheck ./...

workflowcheck:
	@printf $(COLOR) "Run workflow check..."
	@go install go.temporal.io/sdk/contrib/tools/workflowcheck
	@workflowcheck -show-pos ./...

update-sdk:
	$(foreach MOD_FILES_DIR,$(MOD_FILES_DIR),\
		cd $(MOD_FILES_DIR) && \
		go get -u go.temporal.io/sdk@latest && \
		go mod tidy \
	$(NEWLINE))
	
clean:
	rm -rf bin
	
ci-build: staticcheck errcheck workflowcheck bins test
