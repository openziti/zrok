package zrokEdgeSdk

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

func MergeTags(tags *rest_model.Tags, addl map[string]interface{}) *rest_model.Tags {
	for k, v := range addl {
		tags.SubTags[k] = v
	}
	return tags
}