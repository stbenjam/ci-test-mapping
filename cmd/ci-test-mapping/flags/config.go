package flags

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"

	v1 "github.com/openshift-eng/ci-test-mapping/pkg/api/types/v1"
)

type ConfigFlags struct {
	configPath string
}

func NewConfigFlags() *ConfigFlags {
	return &ConfigFlags{
		configPath: "config/openshift-eng.yaml",
	}
}

func (f *ConfigFlags) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&f.configPath,
		"config",
		f.configPath,
		"Location of the configuration",
	)
}

func (f *ConfigFlags) GetConfig() (*v1.Config, error) {
	var config *v1.Config

	// Read the file
	data, err := os.ReadFile(f.configPath)
	if err != nil {
		return config, errors.WithMessage(err, "failed to read config file")
	}

	// Unmarshal the YAML content into the config struct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, errors.WithMessage(err, "failed to unmarshal config")
	}

	return config, nil
}
