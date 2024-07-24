# Component Readiness Test Mapping

Component Readiness needs to know how to map each test to a particular
component and it's capabilities. This tool:

1. Takes a set of metadata about all tests such as it's name and suite
2. Maps the test to exactly one component that provides details about the capabilities that test is testing
3. Writes the result to a json file comitted to this repo in `data/`
4. Pushes the result to BigQuery

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

## Renaming tests

The unfortunate reality is tests may get renamed, so we need to have a
way to compare the test results across renames. To do that, each test
has a stable ID which is the current test name stored in the DB as an
md5sum.

The first stable ID a test has is the one that remains. Component owners are
responsible for ensuring the `StableID` function in their component
returns the same ID for all names of a given test. This can be done with
a simple look-up map, see the monitoring component for an example.

## Removing tests

If a test is removed, or is refactored in such a way (i.e. one to many)
that it's not reasonable to mark them as renames, then it should be
tracked as an obsolete test. In `pkg/obsoletetests` one can manage their
obsolete tests by adding an entry to the set.  For OCP, only staff
engineers can approve a test's removal.

# Test Sources

Currently the tests we use for mapping comes from the corpus of tests
we've previously seen in job results. This list is filtered down to
smaller quantity by selecting only those in certain suites, jobs, or
matching certain names.  This is configurable by specifying a
configuration file. An example is present in
`config/openshift-eng.yaml`.

At a mimimum though, for compatibility with component readiness (and all
other OpenShift tooling), a test must:

* always have a result when it runs, indicating it's success, flake or failure (historically some tests only report failure)

* belong to a test suite

* must have stable names: do not use dynamic names such as full pod names in tests

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
  --mapping-file mapping.json \
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

The BigQuery mapping table may have older entries trimmed, but it should
be assumed to be used in append only mode, so mappings should limit
their results to the most recent entry.

## Syncing with Jira

To create any missing components, run `./ci-test-mapping jira-create`.
You'll need to set the env var `JIRA_TOKEN` to your personal API token
that you can create from your Jira profile page.
