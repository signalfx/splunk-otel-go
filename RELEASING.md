# Release Process

## Bump dependencies

To bump all (execpt the `/build` module's) dependencies run:

```sh
./goyek.sh bump mod
```

## Pre-Release

1. Verify if [wizard](https://app.signalfx.com/#/integrations/go-tracing/description)
   and [official docs](https://help.splunk.com/en/splunk-observability-cloud/manage-data/available-data-sources/supported-integrations-in-splunk-observability-cloud/apm-instrumentation/instrument-a-go-application)
   needs any adjustments.
   Create a Pull Request with documentation updates in
   [splunk/public-o11y-docs](https://github.com/splunk/public-o11y-docs/tree/main/gdi/get-data-in/application/go)
   if necessary.
   Contact @splunk/gdi-docs team if needed.

1. Create a new release branch. I.e. `git checkout -b release-X.X.X main`.

1. Update the version in [`versions.yaml`](versions.yaml)

1. Run the pre-release step which updates `go.mod` and `version.go` files
   in modules for the new release.

    ```sh
    ./goyek.sh prerelease
    ```

1. Merge the branch created by `multimod` into your release branch.

1. Update [CHANGELOG.md](CHANGELOG.md) with new the new release.

1. Push the changes and create a Pull Request on GitHub.

## Tag

Once the Pull Request with all the version changes has been approved
and merged it is time to tag the merged commit.

***IMPORTANT***: It is critical you use the same tag
that you used in the Pre-Release step!
Failure to do so will leave things in a broken state.

***IMPORTANT***:
[There is currently no way to remove an incorrectly tagged version of a Go module](https://github.com/golang/go/issues/34189).
It is critical you make sure the version you push upstream is correct.
[Failure to do so will lead to minor emergencies and tough to work around](https://github.com/open-telemetry/opentelemetry-go/issues/331).

1. Run for the the commit of the merged Pull Request.

    ```sh
    ./goyek.sh -commit <commit> -remote <remote> release
    ```

## Release

Create a Release for the new `<new tag>` on GitHub.
The release body should include all the release notes
for this release taken from [CHANGELOG.md](CHANGELOG.md).

## Post-Release

Bump versions in the following examples:

- [http](https://github.com/signalfx/tracing-examples/tree/main/opentelemetry-tracing/opentelemetry-go/http)
- [lambda](https://github.com/signalfx/tracing-examples/tree/main/opentelemetry-tracing/opentelemetry-lambda/go)
