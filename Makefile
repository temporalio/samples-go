############################# Main targets #############################
# Run all checks, build, and test.
install: clean staticcheck errcheck bins test
########################################################################

##### Variables ######
UNIT_TEST_DIRS := $(sort $(dir $(shell find . -name "*_test.go")))
MAIN_FILES := $(shell find . -name "main.go")
TEST_TIMEOUT := 20s
COLOR := "\e[1;36m%s\e[0m\n"

dir_no_slash = $(patsubst %/,%,$(dir $(1)))
dirname = $(notdir $(call dir_no_slash,$(1)))
parentdirname = $(notdir $(call dir_no_slash,$(call dir_no_slash,$(1))))
define NEWLINE


endef

##### Targets ######
bins:
	@printf $(COLOR) "Build samples..."
	$(foreach MAIN_FILE,$(MAIN_FILES), go build -o bin/$(call parentdirname,$(MAIN_FILE))/$(call dirname,$(MAIN_FILE)) $(MAIN_FILE)$(NEWLINE))

test:
	@printf $(COLOR) "Run unit tests..."
	@rm -f test.log
	$(foreach UNIT_TEST_DIR,$(UNIT_TEST_DIRS),\
		@go test -timeout $(TEST_TIMEOUT) -race $(UNIT_TEST_DIR) | tee -a test.log \
	$(NEWLINE))
	@! grep -q "^--- FAIL" test.log

staticcheck:
	@printf $(COLOR) "Run static check..."
	@GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	@staticcheck ./...

errcheck:
	@printf $(COLOR) "Run error check..."
	@GO111MODULE=off go get -u github.com/kisielk/errcheck
	@errcheck ./...

update-sdk:
	go get -u go.temporal.io/api@master
	go get -u go.temporal.io/sdk@master
	go mod tidy

clean:
	rm -rf bin
