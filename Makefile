# Copyright Splunk Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL := /bin/bash

GO               = go
TIMEOUT          = 15
PKGS             = ./...
BUILD_DIR        = ./build
TEST_RESULTS     = $(CURDIR)/test-results

# Verbose output
V = 0
Q = $(if $(filter 1,$V),,@)

# ALL_MODULES includes ./* dirs (excludes . and ./build dir).
ALL_MODULES := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort )
# All directories with go.mod files related to opentelemetry library. Used for building, testing and linting.
ALL_GO_MOD_DIRS := $(filter-out $(BUILD_DIR), $(ALL_MODULES))
# All directories sub-modules. Used for tagging and generating dependabot config.
SUBMODULES = $(filter-out ., $(ALL_GO_MOD_DIRS))

.DEFAULT_GOAL := goyek
.PHONY: goyek
goyek:
	./goyek.sh

.PHONY: build
build: # build whole codebase
	${call for-all-modules,$(GO) build $(PKGS)}
# Compile all test code.
	${call for-all-modules,$(GO) test -vet=off -run xxxxxMatchNothingxxxxx $(PKGS) >/dev/null}

# Test targets

TEST_TARGETS := test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test tests
test-bench:   ARGS=-run=xxxxxMatchNothingxxxxx -test.benchtime=1ms -bench=.
test-short:   ARGS=-short
test-verbose: ARGS=-v
test-race:    ARGS=-race
$(TEST_TARGETS): test
test tests:
	${call for-all-modules,$(GO) test -timeout $(TIMEOUT)s $(ARGS) $(PKGS)}

# Pre-release targets

.PHONY: add-tag
add-tag: # example usage: make add-tag tag=v1.100.1 commit=<hash>
	$Q [ "$(tag)" ] || ( echo ">> 'tag' is not set"; exit 1 )
	$Q [ "$(commit)" ] || ( echo ">> 'commit' is not set"; exit 1 )
	@echo "Adding tag $(tag)"
	$Q git tag -a $(tag) -s -m "Version $(tag)" $(commit)
	$Q set -e; for dir in $(SUBMODULES); do \
	  (echo Adding tag "$${dir:2}/$(tag)" && \
	 	git tag -a "$${dir:2}/$(tag)" -s -m "Version ${dir:2}/$(tag)" $(commit)); \
	done

.PHONY: delete-tag
delete-tag: # example usage: make delete-tag tag=v1.100.1
	$Q [ "$(tag)" ] || ( echo ">> 'tag' is not set"; exit 1 )
	@echo "Deleting tag $(tag)"
	$Q git tag -d $(tag)
	$Q set -e; for dir in $(SUBMODULES); do \
	  (echo Deleting tag "$${dir:2}/$(tag)" && \
	 	git tag -d "$${dir:2}/$(tag)" ); \
	done

.PHONY: push-tag
push-tag: # example usage: make push-tag remote=origin tag=v1.100.1
	$Q [ "$(remote)" ] || ( echo ">> 'remote' is not set"; exit 1 )
	$Q [ "$(tag)" ] || ( echo ">> 'tag' is not set"; exit 1 )
	@echo "Pushing tag $(tag) to $(remote)"
	$Q git push $(remote) $(tag)
	$Q set -e; for dir in $(SUBMODULES); do \
	  (echo Pushing tag "$${dir:2}/$(tag) to $(remote)" && \
	 	git push $(remote) "$${dir:2}/$(tag)"); \
	done

.PHONY: for-all
for-all: # run a command in all modules, example: make for-all cmd="go mod tidy"
	$Q [ "$(cmd)" ] || ( echo ">> 'cmd' is not set"; exit 1 )
	${call for-all-modules, $(cmd)}

define for-all-modules # run provided command for each module
   $Q EXIT=0 ;\
	for dir in $(ALL_GO_MOD_DIRS); do \
	  echo "${1} in $${dir}"; \
	  (cd "$${dir}" && ${1}) || EXIT=$$?; \
	done ;\
	exit $$EXIT
endef
