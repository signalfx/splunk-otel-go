// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"flag"
	"testing"
)

const (
	argSeparator    = "--"
	argument        = "argument"
	errUnexpectedCI = "unexpected arguments: [ci]"
	flagMessage     = "-msg"
	flagVerbose     = "-v"
	taskCI          = "ci"
	taskRelease     = "release"
)

func TestValidateArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "no arguments",
		},
		{
			name: "tasks only",
			args: []string{taskCI},
		},
		{
			name: "task before flag",
			args: []string{taskCI, flagVerbose},
		},
		{
			name: "task before inline boolean flag",
			args: []string{taskCI, "-v=false"},
		},
		{
			name:    "task after flag",
			args:    []string{flagVerbose, taskCI},
			wantErr: errUnexpectedCI,
		},
		{
			name:    "task after inline boolean flag",
			args:    []string{"-v=false", taskCI},
			wantErr: errUnexpectedCI,
		},
		{
			name:    "argument after boolean flag",
			args:    []string{taskCI, flagVerbose, "false"},
			wantErr: "unexpected arguments: [false]",
		},
		{
			name:    "tasks after flag",
			args:    []string{flagVerbose, "all", "diff"},
			wantErr: "unexpected arguments: [all diff]",
		},
		{
			name:    "task between flag and separator",
			args:    []string{flagVerbose, taskCI, argSeparator, argument},
			wantErr: errUnexpectedCI,
		},
		{
			name: "arguments after separator",
			args: []string{taskCI, flagVerbose, argSeparator, argument, "-other"},
		},
		{
			name: "arguments after separator without tasks",
			args: []string{argSeparator, "all", "diff"},
		},
		{
			name: "separate flag value",
			args: []string{taskRelease, flagMessage, "value", argSeparator, argument},
		},
		{
			name: "separate flag value starting with dash",
			args: []string{taskRelease, flagMessage, "-value"},
		},
		{
			name:    "separator cannot be a flag value",
			args:    []string{taskRelease, flagMessage, argSeparator, argument},
			wantErr: "flag needs an argument: -msg",
		},
		{
			name: "separator as inline flag value",
			args: []string{taskRelease, "-msg=--", argSeparator, argument},
		},
		{
			name:    "unknown flag",
			args:    []string{taskCI, "-unknown"},
			wantErr: "flag provided but not defined: -unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := flag.NewFlagSet("build", flag.ContinueOnError)
			verbose := flags.Bool("v", false, "verbose output")
			message := flags.String("msg", "default", "message")

			err := validateArgs(flags, tt.args)
			if gotErr := errorString(err); gotErr != tt.wantErr {
				t.Errorf("validateArgs(%v) error = %q, want %q", tt.args, gotErr, tt.wantErr)
			}
			if *verbose {
				t.Errorf("validateArgs(%v) modified -v", tt.args)
			}
			if *message != "default" {
				t.Errorf("validateArgs(%v) modified -msg to %q", tt.args, *message)
			}
		})
	}
}

func TestValidateArgsHelp(t *testing.T) {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	if err := validateArgs(flags, []string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Errorf("validateArgs([-h]) error = %v, want flag.ErrHelp", err)
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
