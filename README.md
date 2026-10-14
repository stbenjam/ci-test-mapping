# Component Readiness Test Mapping

Component Readiness needs to know how to map each test to a particular
component and its capabilities. This tool:

1. Takes a set of metadata about all tests such as its name and suite
2. Maps the test to exactly one component that provides details about the capabilities that test is testing
3. Writes the result to a json file comitted to this repo in `data/`
4. Pushes the result to BigQuery

## Updating Mappings

# Overview

Teams own their component code under `pkg/components/<component_name>`
and can handle the mapping as they see fit. New components can copy from
`pkg/components/example` and modify it, or write their own
implementation of the interface. The sample Config interface, which you
can include in your component, provides rich filters on substrings,
sigs, operators, etc.

TRT has made a first pass at assigning ownership to teams, but it's
likely we haven't correctly assigned all tests. Please re-assign tests
to their correct components as needed.  Please also create an OWNERS
file in your directory so your team can manage components and
capabilities without TRT's intervention.

Component owners should return a `TestOwnership` struct from their
identification function. See the details in `pkg/api/v1` for details
about the `TestOwnership` struct.

They should return nil when the test is not theirs.  They should ONLY
return an error on a fatal error such as inability to read from a file.

A test must only map to one component, but may map to several
capabilities.  In the event that two components are vying for a test's
ownership, you may use the `Priority` field in the `TestOwnership`
struct.  The highest value wins.

# Process

These are the general steps for updating a test's ownership:

1. Open the `component.go` for a particular component in
   `pkg/components`. If a test is tied to a namespace, operator, or
   particular SIG, update that section. If you want to match on the
   suite or test name, add a new matcher using one of the existing
   types, such as `Suite`, or `IncludeAny`, which matches a test
   substring.
2. Update the `Priority` field if you need to force a higher or lower
   priority for this component.
3. Run `make mapping`, to apply the new rules to the test names. Review
   the results of the changes by using `git diff`.  Ensure that the
   changes you see are expected.
4. Commit the result and open a pull request on GitHub.

Note: QE mappings (`make mapping-qe`) require BigQuery access to pull
job variant data. Set the `GOOGLE_APPLICATION_CREDENTIALS` environment
variable to your service account credential file before running:

```
export GOOGLE_APPLICATION_CREDENTIALS=~/bq.json
make mapping-qe
```

If you'd like to annotate a test as having additional capabilities,
update `capabilities.go`. A test may have multiple capabilities, but it
can only belong to a single component.

`make mapping` enforces this, as well as disallowing any test to move
back to `Unknown`. If a test is incorrectly attributed, you must find
the owner and move it to their component.

## Renaming tests

The unfortunate reality is tests may get renamed, so we need to have a
way to compare the test results across renames. To do that, each test
has a stable ID which is the current test name stored in the DB as an
md5sum.

The test's first stable ID is the one that remains. Component owners are
responsible for ensuring the `StableID` function in their component
returns the same ID for all names of a given test. This can be done with
a simple look-up map; see the monitoring component for an example.

## Removing tests

If a test is removed, or is refactored in such a way (i.e. one to many)
that it's not reasonable to mark them as renames, then it should be
tracked as an obsolete test. In `pkg/obsoletetests` one can manage their
obsolete tests by adding an entry to the set.  For OCP, only staff
engineers can approve a test's removal.

# Test Sources

Currently the tests used for mapping come from the corpus of tests
previously seen in job results. This list is filtered down to a
smaller quantity by selecting only those in certain suites, jobs, or
matching certain names.  This is configurable by specifying a
configuration file. An example is present in
`config/openshift-eng.yaml`.

At a mimimum though, for compatibility with component readiness (and all
other OpenShift tooling), a test must:

* always have a result when it runs, indicating success, flake, or failure (historically some tests only report failure)

* belong to a test suite

* have a stable name: do not use dynamic names such as full pod names in tests

* have a reasonable way to map to component/capabilities, such as `[sig-XYZ]` present in the test name, and using `[Feature:XYZ]` or `[Capability:XYZ]` to make mapping to capabilities easier

# Usage

See --help for more info.

## Test Mapping

To find unmapped tests, run `make unmapped`.

### Development

For production usage we fetch and push data to BigQuery, but for local
testing you can used locally comitted copies of that data by using
`--mode local`:

```
ci-test-mapping map --mode local
```

### Production

For production, use `--mode bigquery` and provide credentials:

```
ci-test-mapping map --mode bigquery \
  --google-service-account-credential-file ~/bq.json \
  --log-level debug \
  --push-to-bigquery
```

### Alternative data sources/destinations

The BigQuery project, dataset, JUnit table, and component mapping tables
are all configurable.

```
ci-test-mapping map \
    --mode bigquery \
    --bigquery-project openshift-gce-devel \
    --bigquery-dataset ci_analysis_us \
    --table-junit junit \
    --table-mapping component_mapping
```

### Using the BigQuery table for lookups

The BigQuery mapping table should be updated only in append mode (aside from
older entries being trimmed), so mappings should limit their results to the
most recent entry.

## Updating Jira Components

The command `./ci-test-mapping jira-verify` will inform you of any Jira
components that have been removed, and report on any new Jira components
that are available.

To generate local mapping files for any missing components, run
`./ci-test-mapping jira-create`. It updates `pkg/components/` and `pkg/registry/`;
it does not create Jira components or issues.
If any components are being renamed, run `jira-create` first and then
move the configuration to the new component.

Run `./ci-test-mapping jira-verify` to ensure it reports back cleanly.

Configure Jira authentication as described below. Then:

1. Move any configuration for renamed components
2. Delete any obsolete `pkg/components/<component>` directory
3. Remove references to removed components from `pkg/registry`.

### Jira authentication

`map`, `jira-create`, and `jira-verify` share the opt-in flag
`--jira-service-account` (default `false`). Every Jira consumer only reads Jira:

| Command | Jira GET paths | Local or other output |
| --- | --- | --- |
| `map` | `/rest/api/2/project/OCPBUGS/components` | Test and variant mappings; optional BigQuery writes |
| `jira-create` | `/rest/api/2/issue/createmeta/OCPBUGS/issuetypes/` and that path plus the Bug issue-type ID | Component templates and registry entries |
| `jira-verify` | Same creation-metadata paths as `jira-create` | Mapping verification logs |

There are no Jira search consumers or Jira writers. The Makefile's mapping
targets and both mapping invocations in `hack/cronjob-update-mapping.sh` call
`map`. Authentication is independent of `--mode`, `--push-to-bigquery`,
`--map-variant`, and local file generation.

Without the flag, requests continue to use `https://redhat.atlassian.net/`.
`JIRA_TOKEN` takes precedence and supplies `Authorization: Bearer <token>`.
If it is empty or unset, `JIRA_TOKEN_BASIC` supplies `Authorization: Basic
<base64(email:token)>`. For example:

```sh
export JIRA_TOKEN_BASIC=$(printf '%s' 'your_email@redhat.com:the_token' | base64 -w 0)
./ci-test-mapping jira-verify
```

Create the personal API token from your Jira profile. Set only the credential
you intend to use: a nonempty `JIRA_TOKEN` overrides Basic authentication.
`jira-create` and `jira-verify` require legacy credentials; `map` retains its
existing ability to attempt unauthenticated reads. Service-account variables
alone do not enable OAuth and are ignored without the flag.

For service-account authentication, supply a **read-only (RO) OAuth credential**
created in Atlassian Administration, including for `jira-create`:

```sh
export JIRA_CLIENT_ID='<RO credential client ID>'
export JIRA_CLIENT_SECRET='<RO credential client secret>'
export JIRA_CLOUD_ID='<Jira site cloud UUID>'
./ci-test-mapping map --mode=local --jira-service-account
./ci-test-mapping jira-verify --jira-service-account
./ci-test-mapping jira-create --jira-service-account
```

Both `JIRA_CLIENT_ID` and `JIRA_CLIENT_SECRET` must be nonempty, and
`JIRA_CLOUD_ID` must be a UUID. Missing configuration or a failed initial token
grant fails the command; there is no fallback to legacy credentials while the
flag is enabled.

The client sends a form-encoded `client_credentials` grant to
`https://auth.atlassian.com/oauth/token`, acquires a token at startup, and uses
`https://api.atlassian.com/ex/jira/{cloud-id}/` with the same relative REST v2
paths. One HTTP client/token source is reused for all Jira calls and component
retries in a command. `golang.org/x/oauth2/clientcredentials` caches tokens and
obtains replacements on expiry, including its early-expiry margin. No token
files, refresh scheduler, or runtime scopes are used. Keep the supplied command
or job context alive through all calls and renewals. Canceling it stops later
grants; a cached valid token can still serve requests with independent live
contexts. Command context cancellation stops Jira requests and retry waits.
CLI interrupts and SIGTERM retain their default process termination behavior.

Configure scopes on the credential in Atlassian Administration, following
[Atlassian's service-account OAuth guidance](https://support.atlassian.com/user-management/docs/create-oauth-2-0-credential-for-service-accounts/).
The current REST v2 documentation recommends the read scope `read:jira-work`
for [project components](https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-project-components/#api-rest-api-2-project-projectIdOrKey-components-get)
and [creation metadata](https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-issues/#api-rest-api-2-issue-createmeta-projectIdOrKey-issuetypes-get).
The project-components API requires **Browse Projects**. Both creation-metadata
APIs require **Create issues** project permission even though they are GETs;
RO OAuth scopes do not bypass Jira project permissions. Confirm these reads
work for the chosen account before rolling out the metadata commands. No write
OAuth scope is needed by this application.

The endpoints remain documented; local HTTP tests verify gateway routing,
authentication, token reuse/renewal, cancellation, component IDs, both metadata
response shapes, and Documentation filtering. They do not verify live Atlassian
credentials, gateway access, scopes, project permissions, or the site's actual
responses. Existing metadata readers consume a single page; confirm the Bug
issue type and components field are on that page. The components reader expects
an array; projects using Compass may return a paginated shape and require a
separate change. No search API migration is needed.

For a first RO rollout (deployment files are not changed here):

1. Build and deploy the binary containing this change.
2. Inject the RO credential as secret-backed `JIRA_CLIENT_ID` and
   `JIRA_CLIENT_SECRET`, and set `JIRA_CLOUD_ID` to the site's UUID. Ensure both
   credential variables are nonempty. Allow HTTPS egress to `auth.atlassian.com`
   and `api.atlassian.com`.
3. Add `--jira-service-account` to **both** `ci-test-mapping map` invocations in
   the deployed copy of `hack/cronjob-update-mapping.sh` (Engineering and QE),
   or to each equivalent job's command arguments. Environment injection alone
   leaves those invocations using legacy authentication. `prune` and `map-verify`
   do not contact Jira and need no Jira settings.
4. Validate `jira-verify --jira-service-account` and a local `map` run against
   the real site, compare component IDs/results, and check creation-metadata
   permissions before enabling the flag on any metadata jobs. Validate token
   replacement during a run lasting beyond a token lifetime if applicable.
5. Keep the existing legacy credential available during rollout. Roll back by
   removing `--jira-service-account` from the affected command(s); the legacy
   endpoint and Bearer-then-Basic precedence take effect again. The remaining
   service-account variables do not affect legacy runs.

# Component Readiness Variant Mapping

Variant mapping maps a particular job variant to a Jira component. This helps us map 
a column on Component Readiness dashboard to its proper owners. Variant mapping is not
enabled by default. One can use ```--map-variant=true``` (or simply ```--map-variant```) to enable it.

## Add Variant Mapping to components

To add variant mapping to one particular component, all you need to do is to initialize 
Variants slice in your component definition. For example, the following was added to
pkg/components/cloudcompute/vsphereprovider/component.go to map Platform:vsphere to 
component "Cloud Compute / vSphere Provider"

```
		Variants:             []string{"Platform:vsphere"},
```
