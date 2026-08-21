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

// Build is the build pipeline for this repository.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/goyek/goyek/v3"
	"github.com/goyek/x/boot"
)

const (
	dirBuild          = "build"
	exitCodeInvalid   = 2
	repoPackagePrefix = "github.com/signalfx/splunk-otel-go"
)

var flagSkipDocker = flag.Bool("skip-docker", false, "skip tasks and tests using Docker")

func main() {
	if err := os.Chdir(".."); err != nil {
		panic(err)
	}
	goyek.SetDefault(all)
	if err := validateArgs(flag.CommandLine, os.Args[1:]); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(goyek.Output(), err)
		os.Exit(exitCodeInvalid)
	}
	boot.Main()
}

type boolFlag interface {
	IsBoolFlag() bool
}

type validationFlagValue struct {
	isBool bool
}

func (v *validationFlagValue) IsBoolFlag() bool {
	return v.isBool
}

func (*validationFlagValue) Set(string) error {
	return nil
}

func (*validationFlagValue) String() string {
	return ""
}

func validateArgs(flags *flag.FlagSet, args []string) error {
	_, flagArgs := goyek.SplitTasks(args)
	// Only positional arguments after an explicit separator are intentional.
	for i, arg := range flagArgs {
		if arg == "--" {
			flagArgs = flagArgs[:i]
			break
		}
	}

	validationFlags := flag.NewFlagSet(flags.Name(), flag.ContinueOnError)
	validationFlags.SetOutput(io.Discard)
	// Parse no-op copies so validation follows each flag's argument arity
	// without applying flag values before boot.Main parses them.
	flags.VisitAll(func(f *flag.Flag) {
		bf, ok := f.Value.(boolFlag)
		validationFlags.Var(&validationFlagValue{isBool: ok && bf.IsBoolFlag()}, f.Name, f.Usage)
	})

	if err := validationFlags.Parse(flagArgs); err != nil {
		return err
	}
	if validationFlags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", validationFlags.Args())
	}
	return nil
}
