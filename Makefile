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
TIMEOUT          = 60
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
# All directories sub-modules. Used for tagging.
SUBMODULES = $(filter-out ., $(ALL_GO_MOD_DIRS))

.DEFAULT_GOAL := goyek
.PHONY: goyek
goyek:
	./goyek.sh


# Build and test targets

.PHONY: build
build: # build whole codebase
	${call for-all-modules,$(GO) build $(PKGS)}
# Compile all test code.
	${call for-all-modules,$(GO) test -vet=off -run xxxxxMatchNothingxxxxx $(PKGS) >/dev/null}

TEST_TARGETS := test-bench test-short test-verbose test-race
.PHONY: $(TEST_TARGETS) test tests
test-bench:   ARGS=-run=xxxxxMatchNothingxxxxx -test.benchtime=1ms -bench=.
test-short:   ARGS=-short
test-verbose: ARGS=-v
test-race:    ARGS=-race
$(TEST_TARGETS): test
test tests:
	${call for-all-modules,$(GO) test -timeout $(TIMEOUT)s $(ARGS) $(PKGS)}


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
