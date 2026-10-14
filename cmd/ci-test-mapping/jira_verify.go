package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/ci-test-mapping/cmd/ci-test-mapping/flags"
	"github.com/openshift-eng/ci-test-mapping/pkg/registry"
)

var verifyJiraFlags flags.JiraFlags

var verifyCmd = &cobra.Command{
	Use:   "jira-verify",
	Short: "Verify all JIRA components are represented in ci-test-mapping",
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.Info("fetching jira components from jira...")
		jiraClient, err := verifyJiraFlags.GetClient(cmd.Context(), true)
		if err != nil {
			return err
		}
		jiraComponents, err := getJiraComponents(cmd.Context(), jiraClient)
		if err != nil {
			return fmt.Errorf("could not fetch jira components: %w", err)
		}

		logrus.Info("verifying all components have a mapping")
		reg := registry.NewComponentRegistry()
		mappedJiraComponents := sets.New[string]()
		for _, c := range reg.Components {
			mappedJiraComponents.Insert(c.JiraComponents()...)
		}

		for _, component := range jiraComponents {
			if !mappedJiraComponents.Has(component) {
				logrus.WithFields(logrus.Fields{
					"path":    "pkg/" + getPackagePath(component),
					"package": getPackageName(component),
				}).Warningf("no mapping for jira component %q", component)
			}
		}

		jiraComponentsSet := sets.New[string](jiraComponents...)
		for _, mappedComponent := range mappedJiraComponents.UnsortedList() {
			if !jiraComponentsSet.Has(mappedComponent) && mappedComponent != "" {
				logrus.Errorf("unknown component %q not found in jira", mappedComponent)
			}
		}

		logrus.Info("done!")
		return nil
	},
}

func init() {
	verifyJiraFlags.BindFlags(verifyCmd.Flags())
	rootCmd.AddCommand(verifyCmd)
}
