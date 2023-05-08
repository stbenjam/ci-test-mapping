package flags

import (
	"os"

	"github.com/spf13/pflag"
)

// Flags contain auth information for Google BigQuery services.
type Flags struct {
	ServiceAccountCredentialFile string
	OAuthClientCredentialFile    string
}

func NewFlags() *Flags {
	return &Flags{
		OAuthClientCredentialFile: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
	}
}

func (f *Flags) BindFlags(fs *pflag.FlagSet) {
	fs.StringVar(&f.ServiceAccountCredentialFile,
		"google-service-account-credential-file",
		f.ServiceAccountCredentialFile,
		"location of a credential file described by https://cloud.google.com/docs/authentication/production")

	fs.StringVar(&f.OAuthClientCredentialFile,
		"google-oauth-credential-file",
		f.OAuthClientCredentialFile,
		"location of a credential file described by https://developers.google.com/people/quickstart/go, setup from https://cloud.google.com/bigquery/docs/authentication/end-user-installed#client-credentials")
}
