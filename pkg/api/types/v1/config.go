package v1

type Config struct {
	// IncludeSuites is a specific list of suites to include from the tests table.
	IncludeSuites []string `yaml:"includeSuites"`

	// ExcludeSuites is a specific list of suites to exclude from the tests table.
	ExcludeSuites []string `yaml:"excludeSuites"`

	// ExcludeTests is a specific list of tests to exclude from the tests table.
	ExcludeTests []string `yaml:"excludeTests"`

	IncludeJobs []string `yaml:"includeJobs"`

	ExcludeJobs []string `yaml:"excludeJobs"`
}
