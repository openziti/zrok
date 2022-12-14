package zrok_edge_sdk

import (
	"github.com/openziti-test-kitchen/zrok/build"
	"github.com/openziti/edge/rest_model"
)

func ZrokTags() *rest_model.Tags {
	return &rest_model.Tags{
		SubTags: map[string]interface{}{
			"zrok": build.String(),
		},
	}
}

func ZrokServiceTags(svcToken string) *rest_model.Tags {
	tags := ZrokTags()
	tags.SubTags["zrokServiceToken"] = svcToken
	return tags
}
