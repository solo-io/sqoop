package core

import (
	"github.com/solo-io/gloo/pkg/bootstrap"
	"github.com/solo-io/gloo/pkg/bootstrap/configstorage"
	"github.com/pkg/errors"
	qloobootstrap "github.com/solo-io/qloo/pkg/bootstrap"
	"github.com/solo-io/qloo/pkg/configwatcher"
)

type EventLoop struct {}

func Setup(opts bootstrap.Options) (*EventLoop, error) {
	gloo, err := configstorage.Bootstrap(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating gloo client")
	}
	qloo, err := qloobootstrap.Bootstrap(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating qloo client")
	}
	if err := gloo.V1().Register(); err != nil {
		return nil, errors.Wrap(err, "registering gloo client")
	}
	if err := qloo.V1().Register(); err != nil {
		return nil, errors.Wrap(err, "registering qloo storage client")
	}
	cfgWatcher, err := configwatcher.NewConfigWatcher(qloo)
	if err != nil {
		return nil, errors.Wrap(err, "starting watch for QLoo config")
	}
}

func (el *EventLoop) Run(stop <-chan struct{}) {
	for {
		select {

		}
	}
}