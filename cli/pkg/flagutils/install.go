package flagutils

import (
	"github.com/solo-io/sqoop/version"
	"github.com/solo-io/sqoop/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddInstallFlags(set *pflag.FlagSet, install *options.Install) {
	set.BoolVarP(&install.DryRun, "dry-run", "d", false, "Dump the raw installation yaml instead of applying it to kubernetes")
	if !version.IsReleaseVersion() {
		set.StringVar(&install.ReleaseVersion, "release", "", "install using this release version. defaults to the latest github release")
	}
	set.StringVarP(&install.ManifestOverride, "file", "f", "", "Install Gloo from this kubernetes manifest yaml file rather than from a release")
}