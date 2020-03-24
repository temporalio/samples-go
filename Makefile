.PHONY: test bins clean staticcheck errcheck check

# default target
default: check

# all directories with *_test.go files in them
TEST_DIRS := $(sort $(dir $(shell find . -name "*_test.go")))

MAIN_FILES := $(shell find . -name "main.go")

dir_no_slash = $(patsubst %/,%,$(dir $(1)))
dirname = $(notdir $(call dir_no_slash,$(1)))
parentdirname = $(notdir $(call dir_no_slash,$(call dir_no_slash,$(1))))
define NEWLINE


endef

bins:
	@echo Building samples...
	$(foreach MAIN_FILE,$(MAIN_FILES), go build -i -o bin/$(call parentdirname,$(MAIN_FILE))/$(call dirname,$(MAIN_FILE)) $(MAIN_FILE)$(NEWLINE))

test:
	@rm -f test
	@rm -f test.log
	@echo Runing unit tests...
	$(foreach TEST_DIR,$(TEST_DIRS), @go test $(TEST_DIR) | tee -a test.log$(NEWLINE))

staticcheck:
	GO111MODULE=off go get -u honnef.co/go/tools/cmd/staticcheck
	staticcheck ./...

errcheck:
	GO111MODULE=off go get -u github.com/kisielk/errcheck
	errcheck ./...

clean:
	rm -rf bin

check: clean staticcheck errcheck bins test
