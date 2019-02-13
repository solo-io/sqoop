package setup

import (
	"time"

	check "github.com/solo-io/go-checkpoint"

	"github.com/solo-io/gloo/pkg/utils/setuputils"
	"github.com/solo-io/sqoop/pkg/syncer"
	"github.com/solo-io/sqoop/pkg/version"
)

func Main() error {
	start := time.Now()
	check.CallCheck("sqoop", version.Version, start)
	return setuputils.Main(setuputils.SetupOpts{
		SetupFunc:     syncer.Setup,
		ExitOnError:   true,
		LoggingPrefix: "sqoop",
	})
}
