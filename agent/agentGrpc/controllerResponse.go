package agentGrpc

import (
	"strconv"

	"github.com/michaelquigley/df/dd"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ControllerResponse carries a controller answer across the agent RPC boundary.
type ControllerResponse struct {
	Status     int
	Payload    string
	RetryAfter int64
}

// controllerResponseMetadata binds the string-valued fields of the SDK's metadata envelope.
type controllerResponseMetadata struct {
	HTTPStatus string `dd:"http_status,+required"`
	Payload    string `dd:"payload"`
	RetryAfter string `dd:"retry_after"`
}

// AsError preserves the complete diagnostic alongside the controller's classification fields.
func (r ControllerResponse) AsError(cause error) error {
	out := status.New(codes.Unknown, cause.Error())
	values, err := dd.Unbind(&controllerResponseMetadata{
		HTTPStatus: strconv.Itoa(r.Status), Payload: r.Payload,
		RetryAfter: strconv.FormatInt(r.RetryAfter, 10),
	})
	if err != nil {
		return out.Err()
	}
	// adapt the opaque dd map to the SDK's map type without interpreting project fields.
	metadata := make(map[string]string, len(values))
	for key, value := range values {
		text, ok := value.(string)
		if !ok {
			return out.Err()
		}
		metadata[key] = text
	}
	withDetails, err := out.WithDetails(&errdetails.ErrorInfo{
		Domain: "zrok.agent", Reason: "CONTROLLER_RESPONSE",
		Metadata: metadata,
	})
	if err != nil {
		return out.Err()
	}
	return withDetails.Err()
}

// ControllerResponseFromError accepts only the agent's structured controller response detail.
func ControllerResponseFromError(err error) (ControllerResponse, bool) {
	rpc, ok := status.FromError(err)
	if !ok {
		return ControllerResponse{}, false
	}
	for _, detail := range rpc.Details() {
		info, ok := detail.(*errdetails.ErrorInfo)
		if !ok || info.Domain != "zrok.agent" || info.Reason != "CONTROLLER_RESPONSE" {
			continue
		}
		values := make(map[string]any, len(info.Metadata))
		for key, value := range info.Metadata {
			values[key] = value
		}
		metadata, err := dd.New[controllerResponseMetadata](values)
		if err != nil {
			continue
		}
		httpStatus, err := strconv.Atoi(metadata.HTTPStatus)
		if err != nil || httpStatus < 400 || httpStatus >= 600 {
			continue
		}
		retryAfter, _ := strconv.ParseInt(metadata.RetryAfter, 10, 64)
		return ControllerResponse{Status: httpStatus, Payload: metadata.Payload, RetryAfter: retryAfter}, true
	}
	return ControllerResponse{}, false
}
