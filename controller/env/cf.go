package env

import (
	"github.com/michaelquigley/cf"
)

var cfOpts *cf.Options

func GetCfOptions() *cf.Options {
	if cfOpts == nil {
		cfOpts = cf.DefaultOptions()
	}
	return cfOpts
}
