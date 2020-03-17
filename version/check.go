package version

import (
	"github.com/pkg/errors"
	"github.com/solo-io/go-utils/log"
	version "github.com/solo-io/go-utils/versionutils"
)

var (
	glooPkg    = "github.com/solo-io/gloo"
	soloKitPkg = "github.com/solo-io/solo-kit"
)

func CheckVersions() error {
	log.Printf("Checking expected solo kit and gloo versions...")
	tomlTree, err := version.ParseFullToml()
	if err != nil {
		return err
	}

	expectedGlooVersion, err := version.GetDependencyVersionInfo(glooPkg, tomlTree)
	if err != nil {
		return err
	}

	expectedSoloKitVersion, err := version.GetDependencyVersionInfo(soloKitPkg, tomlTree)
	if err != nil {
		return err
	}

	log.Printf("Checking repo versions...")
	actualGlooVersion, err := version.GetGitVersion("../gloo")
	if err != nil {
		return err
	}
	expectedTaggedGlooVersion := version.GetTag(expectedGlooVersion.Version)
	if expectedTaggedGlooVersion != actualGlooVersion {
		return errors.Errorf("Expected gloo version %s, found gloo version %s in repo. Run 'make pin-repos' or fix manually.", expectedTaggedGlooVersion, actualGlooVersion)
	}

	actualSoloKitVersion, err := version.GetGitVersion("../solo-kit")
	if err != nil {
		return err
	}
	expectedTaggedSoloKitVersion := version.GetTag(expectedSoloKitVersion.Version)
	if expectedTaggedSoloKitVersion != actualSoloKitVersion {
		return errors.Errorf("Expected solo kit version %s, found solo kit version %s in repo. Run 'make pin-repos' or fix manually.", expectedTaggedSoloKitVersion, actualSoloKitVersion)
	}
	log.Printf("Versions are pinned correctly.")
	return nil
}
