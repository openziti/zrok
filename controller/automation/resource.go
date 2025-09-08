package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_model"
)

type ResourceManager struct {
	client *Client
}

func NewResourceManager(client *Client) *ResourceManager {
	return &ResourceManager{client: client}
}

type ResourceOptions struct {
	Name       string
	Tags       TagStrategy
	TagContext map[string]interface{}
	Timeout    time.Duration
}

func (ro *ResourceOptions) GetTimeout() time.Duration {
	if ro.Timeout == 0 {
		return 30 * time.Second
	}
	return ro.Timeout
}

func (ro *ResourceOptions) GetTags() *rest_model.Tags {
	if ro.Tags != nil {
		return ro.Tags.GenerateTags(ro.TagContext)
	}
	return &rest_model.Tags{SubTags: make(map[string]interface{})}
}

type FilterOptions struct {
	Filter  string
	Limit   int64
	Offset  int64
	Timeout time.Duration
}

func (fo *FilterOptions) GetTimeout() time.Duration {
	if fo.Timeout == 0 {
		return 30 * time.Second
	}
	return fo.Timeout
}

func (fo *FilterOptions) GetLimit() int64 {
	if fo.Limit == 0 {
		return 100
	}
	return fo.Limit
}

func (rm *ResourceManager) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return rm.client.Edge()
}

func (rm *ResourceManager) Context() context.Context {
	return context.Background()
}

func BuildFilter(field, value string) string {
	return fmt.Sprintf("%s=\"%s\"", field, value)
}

func BuildTagFilter(tag, value string) string {
	return fmt.Sprintf("tags.%s=\"%s\"", tag, value)
}
