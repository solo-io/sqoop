package main

import (
	"github.com/solo-io/go-utils/docsutils"
)

func main() {
	spec := docsutils.DocsPRSpec{
		Owner: "solo-io",
		Repo: "sqoop",
		Product: "sqoop",
		ChangelogPrefix: "sqoop",
		ApiPaths: []string {
			"docs/v1/github.com/solo-io/sqoop",
			"docs/v1/github.com/solo-io/gloo",
			"docs/v1/github.com/solo-io/solo-kit",
			"docs/v1/gogoproto",
			"docs/v1/google",
		},
		CliPrefix: "sqoopctl",
	}
	docsutils.PushDocsCli(&spec)
}