package main

import (
	version "github.com/solo-io/go-utils/versionutils"
	"github.com/solo-io/solo-kit/pkg/utils/log"
)

func main() {
	tomlTree, err := version.ParseToml()
	fatalCheck(err, "parsing error")

	glooVersion, err := version.GetVersion(version.GlooPkg, tomlTree)
	fatalCheck(err, "getting gloo version")

	soloKitVersion, err := version.GetVersion(version.SoloKitPkg, tomlTree)
	fatalCheck(err, "getting solo-kit version")

	fatalCheck(version.PinGitVersion("../gloo", glooVersion), "consider git fetching in gloo repo")

	fatalCheck(version.PinGitVersion("../solo-kit", soloKitVersion), "consider git fetching in solo-kit repo")
}

func fatalCheck(err error, msg string) {
	if err != nil {
		log.Fatalf("Error (%v) unable to pin repos!: %v", msg, err)
	}
}