# Release Process

## Pre-Release

1. Verify if [wizard](https://app.signalfx.com/#/integrations/go-tracing/description)
   and [official docs](https://docs.splunk.com/Observability/gdi/get-data-in/application/go/get-started.html)
   needs any adjustments. Contact @signalfx/gdi-docs team if needed.

2. Run the pre-release script which updates go.mod for submodules to depend on
   the new release. It creates a branch `pre_release_<new tag>`
   that will contain all release changes.

    ```sh
    ./pre_release.sh -t <new tag>
    ```

3. Update [CHANGELOG.md](CHANGELOG.md) with new the new release.

4. Push the changes to upstream and create a Pull Request on GitHub.

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

1. Run on the main branch and specify the commit for the merged Pull Request.

    ```sh
    make add-tag tag=<new tag> commit=<commit>
    ```

2. Push tags to the upstream remote (not your fork): `github.com/signalfx/splunk-otel-go.git`.

    ```sh
    make push-tag tag=<new tag> remote=upstream
    ```

## Release

Create a Release for the new `<new tag>` on GitHub.
The release body should include all the release notes
for this release taken from [CHANGELOG.md](CHANGELOG.md).

## Post-Release

Bump versions in the following examples:

- [gorilla-mux](https://github.com/signalfx/tracing-examples/tree/main/opentelemetry-tracing/opentelemetry-go/gorilla-mux)
- [lambda](https://github.com/signalfx/tracing-examples/tree/main/opentelemetry-tracing/opentelemetry-lambda/go)
