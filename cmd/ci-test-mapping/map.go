package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"

	"cloud.google.com/go/civil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift-eng/ci-test-mapping/cmd/ci-test-mapping/flags"
	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
	"github.com/openshift-eng/ci-test-mapping/pkg/bigquery"
	"github.com/openshift-eng/ci-test-mapping/pkg/components"
	"github.com/openshift-eng/ci-test-mapping/pkg/jira"
	"github.com/openshift-eng/ci-test-mapping/pkg/obsoletetests"
	"github.com/openshift-eng/ci-test-mapping/pkg/registry"
)

const ModeBigQuery = "bigquery"
const ModeLocal = "local"

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map tests to components and capabilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := verifyParams(); err != nil {
			_ = cmd.Usage()
			return err
		}

		var tests []v1.TestInfo
		var tableManager *bigquery.MappingTableManager

		if f.mode == ModeBigQuery {
			// Get a bigquery client
			bigqueryClient, err := bigquery.NewClient(context.Background(),
				f.bigqueryFlags.ServiceAccountCredentialFile,
				f.bigqueryFlags.OAuthClientCredentialFile)
			if err != nil {
				return errors.WithMessage(err, "could not obtain bigquery client")
			}

			// Create or update schema for mapping table
			tableManager = bigquery.NewMappingTableManager(context.Background(), bigqueryClient)
			if err := tableManager.Migrate(); err != nil {
				return errors.WithMessage(err, "could not migrate mapping table")
			}

			// Get a list of all tests from bigquery - this could be swapped out with other
			// mechanisms to get test details later on.
			testLister := bigquery.NewTestTableManager(context.Background(), bigqueryClient)
			tests, err = testLister.ListTests()
			if err != nil {
				return errors.WithMessage(err, "could not list tests")
			}
			if err := writeRecords(tests, "bigquery_tests.json"); err != nil {
				return errors.WithMessage(err, "couldn't write records")
			}
		} else {
			data, err := os.ReadFile(f.testsFile)
			if err != nil {
				return errors.WithMessage(err, "could not fetch tests from file")
			}
			if err := json.Unmarshal(data, &tests); err != nil {
				return errors.WithMessage(err, "could not marshal tests from file")
			}
		}

		// Create a registry of components
		componentRegistry := registry.NewComponentRegistry()

		// Query each component for each test
		now := time.Now()
		createdAt := civil.DateTimeOf(now)
		log.Infof("mapping tests to ownership")

		jiraComponentIDs, err := jira.GetJiraComponents()
		if err != nil {
			return errors.WithMessage(err, "could not get jira component mapping")
		}
		testObsoleter := &obsoletetests.OCPObsoleteTestManager{}
		testIdentifier := components.New(componentRegistry, jiraComponentIDs)
		var newMappings []v1.TestOwnership
		var matched, unmatched int
		success := true
		for i := range tests {
			ownership, err := testIdentifier.Identify(&tests[i])
			if err != nil {
				log.WithError(err).Warningf("encountered error in component identification")
				success = false
				continue
			}
			if ownership != nil {
				if ownership.Component == components.DefaultComponent {
					unmatched++
				} else {
					matched++
				}
				ownership.CreatedAt = createdAt

				ownership.StaffApprovedObsolete = testObsoleter.IsObsolete(&tests[i])
				newMappings = append(newMappings, *ownership)
			}
		}
		if !success {
			return fmt.Errorf("encountered errors while trying to identify tests")
		}

		// Ensure slice is sorted
		sort.Slice(newMappings, func(i, j int) bool {
			return newMappings[i].Name < newMappings[j].Name && newMappings[i].Suite < newMappings[j].Suite
		})

		log.WithFields(log.Fields{
			"matched":   matched,
			"unmatched": unmatched,
		}).Infof("mapping tests to ownership complete in %v", time.Since(now))

		if f.mode == ModeBigQuery && f.pushToBQ {
			now = time.Now()
			log.Infof("pushing to bigquery...")
			if err := tableManager.PushMappings(newMappings); err != nil {
				return errors.WithMessage(err, "could not push records to bigquery")
			}
			log.Infof("push finished in %+v", time.Since(now))
		}

		if err := writeRecords(newMappings, f.mappingFile); err != nil {
			return errors.WithMessage(err, "could not write records to mapping file")
		}
		return nil
	},
}

type MapFlags struct {
	mode          string
	mappingFile   string
	testsFile     string
	pushToBQ      bool
	bigqueryFlags *flags.Flags
}

var f = NewMapFlags()

func NewMapFlags() *MapFlags {
	return &MapFlags{
		bigqueryFlags: flags.NewFlags(),
	}
}

func (f *MapFlags) BindFlags(fs *pflag.FlagSet) {
	f.bigqueryFlags.BindFlags(fs)
}

func init() {
	mapCmd.PersistentFlags().StringVar(&f.mappingFile, "mapping-file", "mapping.json",
		"File containing existing mappings")
	mapCmd.PersistentFlags().StringVar(&f.testsFile, "tests-file", "bigquery_tests.json", "File containing a list of tests to process, see bigquery_tests.json. For local testing without access to canonical test data from BigQuery.")
	mapCmd.PersistentFlags().StringVar(&f.mode, "mode", "local", "Mode (one of: local, bigquery). Local mode doesn't require access to BigQuery and is suitable for local development.")
	mapCmd.PersistentFlags().BoolVar(&f.pushToBQ, "push-to-bigquery", false, "whether or not to push the updated records to bigquery")
	f.BindFlags(mapCmd.Flags())
	rootCmd.AddCommand(mapCmd)
}

func verifyParams() error {
	switch f.mode {
	case ModeBigQuery:
		if f.bigqueryFlags.ServiceAccountCredentialFile == "" && f.bigqueryFlags.OAuthClientCredentialFile == "" {
			return fmt.Errorf("please supply bigquery credentials, or use --mode=local") //nolint
		}
	case ModeLocal:
		if f.pushToBQ {
			return fmt.Errorf("cannot push to bigquery in --mode=local") //nolint
		}

		if f.bigqueryFlags.ServiceAccountCredentialFile != "" || f.bigqueryFlags.OAuthClientCredentialFile != "" {
			return fmt.Errorf("bigquery credentials not required for local mode, maybe you meant to specify --mode=bigquery") //nolint
		}
	default:
		return fmt.Errorf("invalid mode, must be one of: bigquery, local. got: %q", f.mode) //nolint
	}

	return nil
}

func writeRecords(records interface{}, filename string) error {
	now := time.Now()
	log.Infof("writing results to file")
	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.WithError(err).Errorf("could not open file for writing")
		return err
	}
	jsonEncoder := json.NewEncoder(f)
	jsonEncoder.SetIndent("", "  ")

	err = jsonEncoder.Encode(records)
	log.Infof("write complete in %+v", time.Since(now))
	return err
}
