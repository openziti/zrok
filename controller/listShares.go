package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/metrics"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/metadata"
	"github.com/openziti/zrok/util"
	"github.com/pkg/errors"
)

type listSharesHandler struct{}

func newListSharesHandler() *listSharesHandler {
	return &listSharesHandler{}
}

func (h *listSharesHandler) Handle(params metadata.ListSharesParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return metadata.NewListSharesInternalServerError().WithPayload("error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	// build filter from query parameters
	filter := &store.ShareFilter{}

	if params.EnvZID != nil {
		filter.EnvZId = params.EnvZID
	}

	if params.ShareMode != nil {
		filter.ShareMode = params.ShareMode
	}

	if params.BackendMode != nil {
		filter.BackendMode = params.BackendMode
	}

	if params.ShareToken != nil {
		filter.ShareToken = params.ShareToken
	}

	if params.Target != nil {
		filter.Target = params.Target
	}

	if params.PermissionMode != nil {
		filter.PermissionMode = params.PermissionMode
	}

	// parse date filters
	if params.CreatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedAfter)
		if err != nil {
			dl.Errorf("invalid createdAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListSharesBadRequest().WithPayload("invalid createdAfter date format, expected RFC3339")
		}
		filter.CreatedAfter = &t
	}

	if params.CreatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedBefore)
		if err != nil {
			dl.Errorf("invalid createdBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListSharesBadRequest().WithPayload("invalid createdBefore date format, expected RFC3339")
		}
		filter.CreatedBefore = &t
	}

	if params.UpdatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedAfter)
		if err != nil {
			dl.Errorf("invalid updatedAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListSharesBadRequest().WithPayload("invalid updatedAfter date format, expected RFC3339")
		}
		filter.UpdatedAfter = &t
	}

	if params.UpdatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedBefore)
		if err != nil {
			dl.Errorf("invalid updatedBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListSharesBadRequest().WithPayload("invalid updatedBefore date format, expected RFC3339")
		}
		filter.UpdatedBefore = &t
	}

	// query shares with filter
	shares, err := str.FindSharesForAccountWithFilter(int(principal.ID), filter, trx)
	if err != nil {
		dl.Errorf("error finding shares for user '%v': %v", principal.Email, err)
		return metadata.NewListSharesInternalServerError()
	}

	// validate that hasActivity and idle are not both set
	if params.HasActivity != nil && *params.HasActivity && params.Idle != nil && *params.Idle {
		dl.Errorf("hasActivity and idle cannot both be set for user '%v'", principal.Email)
		return metadata.NewListSharesBadRequest().WithPayload("cannot use both hasActivity and idle filters")
	}

	// check for hasActivity or idle filter
	var activeShareIds map[int]bool
	if (params.HasActivity != nil && *params.HasActivity) || (params.Idle != nil && *params.Idle) {
		// parse and validate activity duration
		duration := 24 * time.Hour // default
		if params.ActivityDuration != nil {
			d, err := util.ParseDuration(*params.ActivityDuration)
			if err != nil {
				dl.Errorf("invalid activityDuration ('%v') format for user '%v': %v", *params.ActivityDuration, principal.Email, err)
				return metadata.NewListSharesBadRequest().WithPayload("invalid activityDuration format")
			}
			// validate maximum of 30 days
			if d > 30*24*time.Hour {
				dl.Errorf("activityDuration exceeds maximum for user '%v': %v", principal.Email, d)
				return metadata.NewListSharesBadRequest().WithPayload("activityDuration exceeds maximum of 30d (720h)")
			}
			duration = d
		}

		// query influxdb for active shares
		if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
			activeShareIds, err = findSharesWithActivity(shares, duration, cfg.Metrics.Influx)
			if err != nil {
				dl.Errorf("error querying share activity for user '%v': %v", principal.Email, err)
				// don't fail the request, just log the error and return no activity
				activeShareIds = make(map[int]bool)
			}
		} else {
			// no metrics configured, no shares have activity
			activeShareIds = make(map[int]bool)
		}
	}

	// check account limits
	isLimited := false
	if empty, err := str.IsBandwidthLimitJournalEmpty(int(principal.ID), trx); !empty && err == nil {
		alj, err := str.FindLatestBandwidthLimitJournal(int(principal.ID), trx)
		if err != nil {
			dl.Errorf("error finding account limit journal for '%v': %v", principal.Email, err)
		}
		isLimited = alj != nil && alj.Action == store.LimitLimitAction
	} else if err != nil {
		dl.Errorf("error finding account limit journal for '%v': %v", principal.Email, err)
	}

	// build response
	response := &rest_model_zrok.SharesList{
		Shares: make([]*rest_model_zrok.ShareSummary, 0),
	}

	for _, shr := range shares {
		// apply hasActivity/idle filter
		hasActivity := false
		if activeShareIds != nil {
			hasActivity = activeShareIds[shr.Id]

			// filter by hasActivity (wants active shares)
			if params.HasActivity != nil && *params.HasActivity && !hasActivity {
				continue
			}

			// filter by idle (wants inactive shares)
			if params.Idle != nil && *params.Idle && hasActivity {
				continue
			}
		}

		// get environment z_id
		env, err := str.GetEnvironment(shr.EnvironmentId, trx)
		if err != nil {
			dl.Errorf("error getting environment for share '%v': %v", shr.Token, err)
			continue
		}

		// build frontend endpoints
		frontendEndpoints := buildFrontendEndpointsForShare(shr.Id, shr.Token, shr.FrontendEndpoint, trx)

		// get target
		target := ""
		if shr.BackendProxyEndpoint != nil {
			target = *shr.BackendProxyEndpoint
		}

		summary := &rest_model_zrok.ShareSummary{
			ShareToken:        shr.Token,
			ZID:               shr.ZId,
			EnvZID:            env.ZId,
			ShareMode:         shr.ShareMode,
			BackendMode:       shr.BackendMode,
			FrontendEndpoints: frontendEndpoints,
			Target:            target,
			Limited:           isLimited,
			CreatedAt:         shr.CreatedAt.UnixMilli(),
			UpdatedAt:         shr.UpdatedAt.UnixMilli(),
		}

		response.Shares = append(response.Shares, summary)
	}

	return metadata.NewListSharesOK().WithPayload(response)
}

// findSharesWithActivity queries InfluxDB to find which shares have metrics within the given duration
func findSharesWithActivity(shares []*store.Share, duration time.Duration, influxCfg *metrics.InfluxConfig) (map[int]bool, error) {
	if len(shares) == 0 {
		return make(map[int]bool), nil
	}

	idb := influxdb2.NewClient(influxCfg.Url, influxCfg.Token)
	defer idb.Close()
	queryApi := idb.QueryAPI(influxCfg.Org)

	// build filter for share tokens
	shareFilter := "|> filter(fn: (r) =>"
	for i, shr := range shares {
		if i > 0 {
			shareFilter += " or"
		}
		shareFilter += fmt.Sprintf(" r[\"share\"] == \"%s\"", shr.Token)
	}
	shareFilter += ")"

	query := fmt.Sprintf("from(bucket: \"%v\")\n", influxCfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		shareFilter + "\n" +
		"|> group(columns: [\"share\"])\n" +
		"|> sum()"

	result, err := queryApi.Query(context.Background(), query)
	if err != nil {
		return nil, errors.Wrap(err, "error querying influxdb for share activity")
	}

	// build map of share token to share id
	tokenToId := make(map[string]int)
	for _, shr := range shares {
		tokenToId[shr.Token] = shr.Id
	}

	activeShareIds := make(map[int]bool)
	for result.Next() {
		shareToken, ok := result.Record().ValueByKey("share").(string)
		if !ok {
			continue
		}
		// any non-zero value means there was activity
		if val, ok := result.Record().Value().(int64); ok && val > 0 {
			if shareId, found := tokenToId[shareToken]; found {
				activeShareIds[shareId] = true
			}
		}
	}

	if result.Err() != nil {
		return nil, errors.Wrap(result.Err(), "error reading influxdb query results")
	}

	return activeShareIds, nil
}
