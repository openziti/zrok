package automation

import (
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/build"
)

type TagStrategy interface {
	GenerateTags(context map[string]interface{}) *rest_model.Tags
}

type BaseTagStrategy struct {
	baseTags map[string]interface{}
}

func NewBaseTagStrategy(baseTags map[string]interface{}) *BaseTagStrategy {
	return &BaseTagStrategy{baseTags: baseTags}
}

func (bts *BaseTagStrategy) GenerateTags(context map[string]interface{}) *rest_model.Tags {
	tags := &rest_model.Tags{
		SubTags: make(map[string]interface{}),
	}

	// add base tags first
	for k, v := range bts.baseTags {
		tags.SubTags[k] = v
	}

	// add context tags (can override base tags)
	for k, v := range context {
		tags.SubTags[k] = v
	}

	return tags
}

type SimpleTagStrategy struct {
	tags map[string]interface{}
}

func NewSimpleTagStrategy(tags map[string]interface{}) *SimpleTagStrategy {
	return &SimpleTagStrategy{tags: tags}
}

func (sts *SimpleTagStrategy) GenerateTags(context map[string]interface{}) *rest_model.Tags {
	tags := &rest_model.Tags{
		SubTags: make(map[string]interface{}),
	}

	for k, v := range sts.tags {
		tags.SubTags[k] = v
	}

	return tags
}

func MergeTags(base *rest_model.Tags, additional map[string]interface{}) *rest_model.Tags {
	if base == nil {
		base = &rest_model.Tags{SubTags: make(map[string]interface{})}
	}
	if base.SubTags == nil {
		base.SubTags = make(map[string]interface{})
	}

	for k, v := range additional {
		base.SubTags[k] = v
	}
	return base
}

// zrok-specific tag strategies

type ZrokTagStrategy struct {
	additionalTags map[string]interface{}
}

func NewZrokTagStrategy() *ZrokTagStrategy {
	return &ZrokTagStrategy{
		additionalTags: make(map[string]interface{}),
	}
}

func NewZrokTagStrategyWithTags(tags map[string]interface{}) *ZrokTagStrategy {
	zts := &ZrokTagStrategy{
		additionalTags: make(map[string]interface{}),
	}
	for k, v := range tags {
		zts.additionalTags[k] = v
	}
	return zts
}

func (zts *ZrokTagStrategy) WithTag(key string, value interface{}) *ZrokTagStrategy {
	zts.additionalTags[key] = value
	return zts
}

func (zts *ZrokTagStrategy) WithShareToken(token string) *ZrokTagStrategy {
	zts.additionalTags["zrokShareToken"] = token
	return zts
}

func (zts *ZrokTagStrategy) WithAgentRemote(enrollmentToken, envZId string) *ZrokTagStrategy {
	zts.additionalTags["zrokAgentRemote"] = enrollmentToken
	zts.additionalTags["zrokEnvZId"] = envZId
	return zts
}

func (zts *ZrokTagStrategy) GenerateTags(context map[string]interface{}) *rest_model.Tags {
	tags := &rest_model.Tags{
		SubTags: make(map[string]interface{}),
	}

	// always include the zrok build tag
	tags.SubTags["zrok"] = build.String()

	// add additional tags
	for k, v := range zts.additionalTags {
		tags.SubTags[k] = v
	}

	// add context tags (can override additional tags)
	for k, v := range context {
		tags.SubTags[k] = v
	}

	return tags
}

// convenience functions for common zrok tag patterns

func ZrokShareTags(shareToken string) TagStrategy {
	return NewZrokTagStrategy().WithShareToken(shareToken)
}

func ZrokAgentRemoteTags(enrollmentToken, envZId string) TagStrategy {
	return NewZrokTagStrategy().WithAgentRemote(enrollmentToken, envZId)
}

func ZrokBaseTags() TagStrategy {
	return NewZrokTagStrategy()
}
