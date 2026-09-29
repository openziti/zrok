package automation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/edge-api/rest_management_api_client"
	"github.com/openziti/edge-api/rest_model"
	"github.com/pkg/errors"
)

type IResourceManager[T any, CreateOpts any] interface {
	Create(opts CreateOpts) (string, error)
	Delete(id string) error
	Find(opts *FilterOptions) ([]*T, error)
	GetByID(id string) (*T, error)
	GetByName(name string) (*T, error)
}

type ErrorType int

const (
	ErrorTypeNotFound ErrorType = iota
	ErrorTypePermission
	ErrorTypeValidation
	ErrorTypeNetwork
	ErrorTypeInternal
)

type AutomationError struct {
	Type      ErrorType
	Resource  string
	Operation string
	Cause     error
}

func (e *AutomationError) Error() string {
	return fmt.Sprintf("%s operation failed on %s: %v", e.Operation, e.Resource, e.Cause)
}

func (e *AutomationError) IsNotFound() bool {
	return e.Type == ErrorTypeNotFound
}

func (e *AutomationError) IsRetryable() bool {
	return e.Type == ErrorTypeNetwork
}

func NewNotFoundError(resource, operation string, cause error) *AutomationError {
	return &AutomationError{
		Type:      ErrorTypeNotFound,
		Resource:  resource,
		Operation: operation,
		Cause:     cause,
	}
}

// edgeErrorPayload is implemented by every generated edge-api error response type.
type edgeErrorPayload interface {
	GetPayload() *rest_model.APIErrorEnvelope
}

// wrapEdgeError wraps a failed edge-api call and, when the response carries Ziti's error
// envelope, appends its code and message. the generated error stays in the chain so
// IsNotFound and errors.As keep working.
func wrapEdgeError(err error, format string, args ...interface{}) error {
	var generated edgeErrorPayload
	if !errors.As(err, &generated) {
		return errors.Wrapf(err, format, args...)
	}
	envelope := generated.GetPayload()
	if envelope == nil || envelope.Error == nil {
		return errors.Wrapf(err, format, args...)
	}
	apiErr := envelope.Error
	msg := fmt.Sprintf(format, args...) + ": ziti " + apiErr.Code + ": " + apiErr.Message
	if cause := edgeErrorCause(apiErr); cause != "" {
		msg += " (cause: " + cause + ")"
	}
	// the generated text is '[METHOD /path][status] operation  payload'; keep the part before the
	// payload, which formats as a pointer, so the line still names the operation and status.
	if operation, _, ok := strings.Cut(generated.(error).Error(), "  "); ok {
		msg += " " + operation
	}
	return &edgeError{msg: msg, cause: err}
}

func edgeErrorCause(apiErr *rest_model.APIError) string {
	if apiErr.CauseMessage != "" {
		return apiErr.CauseMessage
	}
	if cause := apiErr.Cause; cause != nil {
		if cause.APIError.Message != "" {
			return cause.APIError.Message
		}
		if cause.Reason != "" {
			if cause.Field != "" {
				return cause.Field + ": " + cause.Reason
			}
			return cause.Reason
		}
	}
	return ""
}

type edgeError struct {
	msg   string
	cause error
}

func (e *edgeError) Error() string { return e.msg }
func (e *edgeError) Unwrap() error { return e.cause }
func (e *edgeError) Cause() error  { return e.cause }

type BaseOptions struct {
	Name       string
	Tags       *Tags
	TagContext map[string]interface{}
	Timeout    time.Duration
}

func (bo *BaseOptions) GetTimeout() time.Duration {
	if bo.Timeout == 0 {
		return DefaultRequestTimeout
	}
	return bo.Timeout
}

func (bo *BaseOptions) GetTags() *rest_model.Tags {
	if bo.Tags != nil {
		return bo.Tags.ToRestModel()
	}
	return &rest_model.Tags{SubTags: make(map[string]interface{})}
}

type BaseResourceManager[T any] struct {
	client *ZitiAutomation
}

func NewBaseResourceManager[T any](ziti *ZitiAutomation) *BaseResourceManager[T] {
	return &BaseResourceManager[T]{client: ziti}
}

func (brm *BaseResourceManager[T]) Edge() *rest_management_api_client.ZitiEdgeManagement {
	return brm.client.Edge()
}

func (brm *BaseResourceManager[T]) Context() context.Context {
	return context.Background()
}

// generic helper for GetByID operations
func GetByID[T any](finder func(*FilterOptions) ([]*T, error), id string, resourceType string) (*T, error) {
	opts := &FilterOptions{Filter: BuildFilter("id", id)}
	items, err := finder(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding %s by id", resourceType)
	}
	if len(items) == 0 {
		return nil, NewNotFoundError(resourceType, "GetByID", errors.Errorf("%s with id '%s' not found", resourceType, id))
	}
	if len(items) != 1 {
		return nil, errors.Errorf("expected 1 %s, found %d", resourceType, len(items))
	}
	return items[0], nil
}

// generic helper for GetByName operations
func GetByName[T any](finder func(*FilterOptions) ([]*T, error), name string, resourceType string) (*T, error) {
	opts := &FilterOptions{Filter: BuildFilter("name", name)}
	items, err := finder(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding %s by name", resourceType)
	}
	if len(items) == 0 {
		return nil, NewNotFoundError(resourceType, "GetByName", errors.Errorf("%s with name '%s' not found", resourceType, name))
	}
	if len(items) != 1 {
		return nil, errors.Errorf("expected 1 %s with name '%s', found %d", resourceType, name, len(items))
	}
	return items[0], nil
}

// generic helper for bulk delete operations
func DeleteWithFilter[T any](finder func(*FilterOptions) ([]*T, error), deleter func(string) error, filter string, resourceType string) error {
	opts := &FilterOptions{Filter: filter}
	items, err := finder(opts)
	if err != nil {
		return errors.Wrapf(err, "error finding %s for deletion", resourceType)
	}

	for _, item := range items {
		id, err := getResourceID(item)
		if err != nil {
			return errors.Wrapf(err, "error extracting ID for %s", resourceType)
		}
		if err := deleter(id); err != nil {
			// an object that vanished between the listing and its delete is already where we want it.
			if IsNotFound(err) {
				dl.Debugf("%s '%s' already deleted", resourceType, id)
				continue
			}
			return errors.Wrapf(err, "error deleting %s '%s'", resourceType, id)
		}
	}

	return nil
}

// helper to extract ID from any resource type
func getResourceID(item interface{}) (string, error) {
	switch v := item.(type) {
	case *rest_model.IdentityDetail:
		return *v.ID, nil
	case *rest_model.ServiceDetail:
		return *v.ID, nil
	case *rest_model.ConfigDetail:
		return *v.ID, nil
	case *rest_model.ConfigTypeDetail:
		return *v.ID, nil
	case *rest_model.EdgeRouterPolicyDetail:
		return *v.ID, nil
	case *rest_model.ServiceEdgeRouterPolicyDetail:
		return *v.ID, nil
	case *rest_model.ServicePolicyDetail:
		return *v.ID, nil
	default:
		return "", errors.Errorf("unsupported resource type: %T", item)
	}
}

type FilterOptions struct {
	Filter  string
	Limit   int64
	Offset  int64
	Timeout time.Duration
}

func (fo *FilterOptions) GetTimeout() time.Duration {
	if fo.Timeout == 0 {
		return DefaultRequestTimeout
	}
	return fo.Timeout
}

func (fo *FilterOptions) GetLimit() int64 {
	if fo.Limit == 0 {
		return 100
	}
	return fo.Limit
}

func BuildFilter(field, value string) string {
	return fmt.Sprintf("%s=\"%s\"", field, value)
}

func BuildTagFilter(tag, value string) string {
	return fmt.Sprintf("tags.%s=\"%s\"", tag, value)
}
