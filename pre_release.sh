#!/bin/bash

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

# Script which helps preparing a new version for all Go Modules.
# Based on: https://github.com/open-telemetry/opentelemetry-go/blob/main/pre_release.sh

readonly DEFAULT_BRANCH=main

set -e

help()
{
   printf "\n"
   printf "Usage: $0 -t tag\n"
   printf "\t-t Unreleased tag. Update all go.mod with this tag.\n"
   exit 1 # Exit script after printing help
}

while getopts "t:" opt
do
   case "$opt" in
      t ) TAG="$OPTARG" ;;
      ? ) help ;; # Print help
   esac
done

# Print help in case parameters are empty
if [ -z "$TAG" ]
then
   printf "Tag is missing\n";
   help
fi

# Validate semver
SEMVER_REGEX="^v(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
if [[ "${TAG}" =~ ${SEMVER_REGEX} ]]; then
	printf "${TAG} is valid semver tag.\n"
else
	printf "${TAG} is not a valid semver tag.\n"
	exit -1
fi

TAG_FOUND=`git tag --list ${TAG}`
if [[ ${TAG_FOUND} = ${TAG} ]] ; then
        printf "Tag ${TAG} already exists\n"
        exit -1
fi

# Check if there is no WIP in Git
cd $(dirname $0)

if ! git diff --quiet; then \
	printf "Working tree is not clean, can't proceed with the release process\n"
	git status
	git diff
	exit 1
fi

# Prepare new branch
git checkout -b pre_release_${TAG} "$DEFAULT_BRANCH"

# Update version.go
VERSION_IN_FILE=$(echo "${TAG}" | grep -o '^v[0-9]\+\.[0-9]\+\.[0-9]\+')
VERSION_IN_FILE="${VERSION_IN_FILE#v}" # Strip leading v
cp ./version.go ./version.go.bak
sed "s/\( \"\)[0-9]*\.[0-9]*\.[0-9]*\"/\1${VERSION_IN_FILE}\"/" ./version.go.bak >./version.go
rm -f ./version.go.bak

# Update go.mod files
PACKAGE_DIRS=$(find . -mindepth 2 -type f -name 'go.mod' -exec dirname {} \; | egrep -v 'tools' | sed 's/^\.\///' | sort)

for dir in $PACKAGE_DIRS; do
	cp "${dir}/go.mod" "${dir}/go.mod.bak"
	sed "s/github.com\/signalfx\/splunk-otel-go\([^ ]*\) v[0-9]*\.[0-9]*\.[0-9]/github.com\/signalfx\/splunk-otel-go\1 ${TAG}/" "${dir}/go.mod.bak" >"${dir}/go.mod"
	rm -f "${dir}/go.mod.bak"
done

printf "Updating go.sum files\n"
./goyek.sh mod-tidy

# Add changes and commit.
git add .
git commit -m "Release $TAG"

printf "Now push changes and create a pull request\n"
