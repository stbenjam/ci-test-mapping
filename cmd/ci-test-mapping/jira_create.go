package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/openshift-eng/ci-test-mapping/cmd/ci-test-mapping/flags"
	"github.com/openshift-eng/ci-test-mapping/pkg/jira"
	"github.com/openshift-eng/ci-test-mapping/pkg/registry"
)

const createMetadataPath = "rest/api/2/issue/createmeta/OCPBUGS/issuetypes/"

var createJiraFlags flags.JiraFlags

var createCmd = &cobra.Command{
	Use:   "jira-create",
	Short: "Create mapping components for missing Jira components",
	RunE: func(cmd *cobra.Command, args []string) error {
		jiraClient, err := createJiraFlags.GetClient(cmd.Context(), true)
		if err != nil {
			return err
		}

		components, err := getJiraComponents(cmd.Context(), jiraClient)
		if err != nil {
			return fmt.Errorf("could not fetch jira components: %w", err)
		}

		reg := registry.NewComponentRegistry()
		knownJiraComponents := sets.New[string]()
		for _, c := range reg.Components {
			knownJiraComponents.Insert(c.JiraComponents()...)
		}

		for _, component := range components {
			if !knownJiraComponents.Has(component) {
				logrus.WithFields(logrus.Fields{
					"path":    "pkg/" + getPackagePath(component),
					"package": getPackageName(component),
				}).Infof("no mapping for jira component %q, creating...", component)
				if err := copyTemplate(component); err != nil {
					return fmt.Errorf("couldn't copy template: %w", err)
				}
			}
		}
		return nil
	},
}

func getJiraBugTypeID(ctx context.Context, client *jira.Client) (string, error) {
	req, err := client.NewRequest(ctx, createMetadataPath)
	if err != nil {
		return "", fmt.Errorf("could not create GET client: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error while reading types response: %w", err)
	}

	type JiraTypes struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	var jiraTypes struct {
		Values     []JiraTypes `json:"values"`
		IssueTypes []JiraTypes `json:"issueTypes"`
	}

	if err := json.Unmarshal(body, &jiraTypes); err != nil {
		return "", err
	}

	types := jiraTypes.IssueTypes
	if len(types) == 0 {
		types = jiraTypes.Values
	}

	for _, value := range types {
		if value.Name == "Bug" {
			return value.ID, nil
		}
	}

	return "", nil
}

func getJiraComponents(ctx context.Context, client *jira.Client) ([]string, error) {

	// bug type ids are not constant across environments so look it up
	id, err := getJiraBugTypeID(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("could not fetch jira bug type: %w", err)
	}

	if len(id) == 0 {
		return nil, fmt.Errorf("jira bug type id required")
	}

	componentsPath := createMetadataPath + id
	req, err := client.NewRequest(ctx, componentsPath)
	if err != nil {
		return nil, fmt.Errorf("could not create GET client: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading response: %w", err)
	}

	type Fields struct {
		FieldID       string `json:"fieldId"`
		AllowedValues []struct {
			Name string `json:"name"`
		} `json:"allowedValues"`
	}

	var jiraComponents struct {
		Values []Fields `json:"values"`
		Fields []Fields `json:"fields"`
	}
	if err := json.Unmarshal(body, &jiraComponents); err != nil {
		return nil, err
	}

	// atlassian cloud response is slightly different
	// handle both
	fields := jiraComponents.Fields
	if len(fields) == 0 {
		fields = jiraComponents.Values
	}

	var components []string
	for _, value := range fields {
		if value.FieldID == "components" {
			for _, allowedValue := range value.AllowedValues {
				if strings.Contains(allowedValue.Name, "Documentation") {
					continue
				}

				components = append(components, allowedValue.Name)
			}
		}
	}

	return components, nil
}

func getPackagePath(input string) string {
	parts := strings.Split(input, "/")
	for i := range parts {
		parts[i] = strings.ToLower(getComponentName(parts[i]))
	}
	return strings.Join(parts, "/")
}

func getPackageName(input string) string {
	re := regexp.MustCompile(`\([^)]*\)`)
	input = re.ReplaceAllString(input, "")
	input = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, input)
	return strings.ToLower(input)
}

func getComponentName(input string) string {
	// Remove everything inside parentheses using a regular expression
	re := regexp.MustCompile(`\([^)]*\)`)
	input = re.ReplaceAllString(input, "")

	// Replace spaces and dashes with newlines
	input = strings.ReplaceAll(input, " ", "\n")
	input = strings.ReplaceAll(input, "-", "\n")

	// Convert the first letter of each word to uppercase
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = strings.ToUpper(string(line[0])) + line[1:]
	}
	input = strings.Join(lines, "")

	// Remove trailing space
	input = strings.TrimRight(input, " ")

	return input
}

func copyTemplate(component string) error {
	destPath := "./pkg/components/" + getPackagePath(component)
	srcPath := "./pkg/components/example"

	files, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}
	name := getComponentName(component)
	parts := strings.Split(name, "/")
	if len(parts) > 1 {
		name = parts[len(parts)-1]
	}
	for _, f := range files {
		src := srcPath + "/" + f.Name()
		dest := destPath + "/" + f.Name()
		dir := filepath.Dir(dest)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		// Read the source file
		content, err := os.ReadFile(src)
		if err != nil {
			return err
		}

		newContent := strings.ReplaceAll(string(content),
			"ExampleComponent",
			fmt.Sprintf("%sComponent", name))
		newContent = strings.ReplaceAll(newContent,
			"package example",
			fmt.Sprintf("package %s", getPackageName(component)))
		newContent = strings.ReplaceAll(newContent,
			"Example",
			component)

		err = os.WriteFile(dest, []byte(newContent), 0o644) //nolint:gosec
		if err != nil {
			return err
		}

	}
	// Read the source file
	regFile, err := os.ReadFile("./pkg/registry/registry.go")
	if err != nil {
		return err
	}

	prepend := "// New components go here"
	importString := fmt.Sprintf(`"github.com/openshift-eng/ci-test-mapping/pkg/components/%s"`,
		getPackagePath(component))

	registerCmd :=
		fmt.Sprintf(`r.Register("%s", &%s.%s)`,
			component,
			getPackageName(component),
			fmt.Sprintf("%sComponent", name),
		)
	newContent := strings.ReplaceAll(string(regFile),
		prepend,
		fmt.Sprintf("%s\n\t%s", registerCmd, prepend))
	newContent = strings.ReplaceAll(newContent,
		"import (",
		fmt.Sprintf("import (\n\t%s", importString))

	err = os.WriteFile("./pkg/registry/registry.go", []byte(newContent), 0o644) //nolint:gosec
	if err != nil {
		return err
	}
	return nil
}

func init() {
	createJiraFlags.BindFlags(createCmd.Flags())
	rootCmd.AddCommand(createCmd)
}
