package v1

type Config struct {
	// IncludeSuites is a specific list of suites to include from the tests table.
	IncludeSuites []string `yaml:"includeSuites"`

	// ExcludeSuites is a specific list of suites to exclude from the tests table.
	ExcludeSuites []string `yaml:"excludeSuites"`

	// ExcludeTests is a specific list of tests to exclude from the tests table.
	ExcludeTests []string `yaml:"excludeTests"`

	// IncludeJobs is a specific list of CI jobs to include from the tests table.
	IncludeJobs []string `yaml:"includeJobs"`

	// ExcludeJobs is a specific list of CI jobs to exclude from the tests table.
	ExcludeJobs []string `yaml:"excludeJobs"`
}
