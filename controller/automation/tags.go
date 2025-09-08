package automation

import "github.com/openziti/edge-api/rest_model"

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
