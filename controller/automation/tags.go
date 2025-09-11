package automation

import (
	"github.com/openziti/edge-api/rest_model"
	"github.com/openziti/zrok/build"
)

type Tags struct {
	tags map[string]interface{}
}

func NewTags() *Tags {
	return &Tags{
		tags: make(map[string]interface{}),
	}
}

func (t *Tags) WithTag(key string, value interface{}) *Tags {
	t.tags[key] = value
	return t
}

func (t *Tags) WithZrok() *Tags {
	t.tags["zrok"] = build.String()
	return t
}

func (t *Tags) WithEmail(email string) *Tags {
	t.tags["zrokEmail"] = email
	return t
}

func (t *Tags) WithShareToken(token string) *Tags {
	t.tags["zrokShareToken"] = token
	return t
}

func (t *Tags) WithAgentRemote(enrollmentToken, envZId string) *Tags {
	t.tags["zrokAgentRemote"] = enrollmentToken
	t.tags["zrokEnvZId"] = envZId
	return t
}

func (t *Tags) ToRestModel() *rest_model.Tags {
	return &rest_model.Tags{
		SubTags: t.tags,
	}
}

// convenience functions for common patterns

func ZrokTags() *Tags {
	return NewTags().WithZrok()
}

func ZrokShareTags(shareToken string) *Tags {
	return NewTags().WithZrok().WithShareToken(shareToken)
}

func ZrokAgentRemoteTags(enrollmentToken, envZId string) *Tags {
	return NewTags().WithZrok().WithAgentRemote(enrollmentToken, envZId)
}
