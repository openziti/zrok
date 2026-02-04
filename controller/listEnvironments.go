package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/runtime/middleware"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/metrics"
	"github.com/openziti/zrok/v2/controller/store"
	"github.com/openziti/zrok/v2/rest_model_zrok"
	"github.com/openziti/zrok/v2/rest_server_zrok/operations/metadata"
	"github.com/openziti/zrok/v2/util"
	"github.com/pkg/errors"
)

type listEnvironmentsHandler struct{}

func newListEnvironmentsHandler() *listEnvironmentsHandler {
	return &listEnvironmentsHandler{}
}

func (h *listEnvironmentsHandler) Handle(params metadata.ListEnvironmentsParams, principal *rest_model_zrok.Principal) middleware.Responder {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return metadata.NewListEnvironmentsInternalServerError().WithPayload("error starting transaction")
	}
	defer func() { _ = trx.Rollback() }()

	// build filter from query parameters
	filter := &store.EnvironmentFilter{}

	if params.Description != nil {
		filter.Description = params.Description
	}

	if params.Host != nil {
		filter.Host = params.Host
	}

	if params.Address != nil {
		filter.Address = params.Address
	}

	if params.ShareCount != nil {
		filter.ShareCount = params.ShareCount
	}

	if params.AccessCount != nil {
		filter.AccessCount = params.AccessCount
	}

	if params.HasShares != nil {
		filter.HasShares = params.HasShares
	}

	if params.HasAccesses != nil {
		filter.HasAccesses = params.HasAccesses
	}

	// parse date filters
	if params.CreatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedAfter)
		if err != nil {
			dl.Errorf("invalid createdAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListEnvironmentsBadRequest().WithPayload("invalid createdAfter date format, expected RFC3339")
		}
		filter.CreatedAfter = &t
	}

	if params.CreatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.CreatedBefore)
		if err != nil {
			dl.Errorf("invalid createdBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListEnvironmentsBadRequest().WithPayload("invalid createdBefore date format, expected RFC3339")
		}
		filter.CreatedBefore = &t
	}

	if params.UpdatedAfter != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedAfter)
		if err != nil {
			dl.Errorf("invalid updatedAfter format for user '%v': %v", principal.Email, err)
			return metadata.NewListEnvironmentsBadRequest().WithPayload("invalid updatedAfter date format, expected RFC3339")
		}
		filter.UpdatedAfter = &t
	}

	if params.UpdatedBefore != nil {
		t, err := time.Parse(time.RFC3339, *params.UpdatedBefore)
		if err != nil {
			dl.Errorf("invalid updatedBefore format for user '%v': %v", principal.Email, err)
			return metadata.NewListEnvironmentsBadRequest().WithPayload("invalid updatedBefore date format, expected RFC3339")
		}
		filter.UpdatedBefore = &t
	}

	// query environments with filter
	envs, err := str.FindEnvironmentsForAccountWithFilter(int(principal.ID), filter, trx)
	if err != nil {
		dl.Errorf("error finding environments for user '%v': %v", principal.Email, err)
		return metadata.NewListEnvironmentsInternalServerError()
	}

	// validate that hasActivity and idle are not both set
	if params.HasActivity != nil && *params.HasActivity && params.Idle != nil && *params.Idle {
		dl.Errorf("hasActivity and idle cannot both be set for user '%v'", principal.Email)
		return metadata.NewListEnvironmentsBadRequest().WithPayload("cannot use both hasActivity and idle filters")
	}

	// check for hasActivity or idle filter
	var activeEnvIds map[int]bool
	if (params.HasActivity != nil && *params.HasActivity) || (params.Idle != nil && *params.Idle) {
		// parse and validate activity duration
		duration := 24 * time.Hour // default
		if params.ActivityDuration != nil {
			d, err := util.ParseDuration(*params.ActivityDuration)
			if err != nil {
				dl.Errorf("invalid activityDuration ('%v') format for user '%v': %v", *params.ActivityDuration, principal.Email, err)
				return metadata.NewListEnvironmentsBadRequest().WithPayload("invalid activityDuration format")
			}
			// validate maximum of 30 days
			if d > 30*24*time.Hour {
				dl.Errorf("activityDuration exceeds maximum for user '%v': %v", principal.Email, d)
				return metadata.NewListEnvironmentsBadRequest().WithPayload("activityDuration exceeds maximum of 30d (720h)")
			}
			duration = d
		}

		// query influxdb for active environments
		if cfg.Metrics != nil && cfg.Metrics.Influx != nil {
			activeEnvIds, err = findEnvironmentsWithActivity(envs, duration, cfg.Metrics.Influx)
			if err != nil {
				dl.Errorf("error querying environment activity for user '%v': %v", principal.Email, err)
				// don't fail the request, just log the error and return no activity
				activeEnvIds = make(map[int]bool)
			}
		} else {
			// no metrics configured, no environments have activity
			activeEnvIds = make(map[int]bool)
		}
	}

	// check for remoteAgent filter
	var agentEnvIds map[int]bool
	if params.RemoteAgent != nil {
		agentEnvIds, err = findEnvironmentsWithAgents(envs, trx)
		if err != nil {
			dl.Errorf("error checking remote agents for user '%v': %v", principal.Email, err)
			return metadata.NewListEnvironmentsInternalServerError()
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
	response := &rest_model_zrok.EnvironmentsList{
		Environments: make([]*rest_model_zrok.EnvironmentSummary, 0),
	}

	for _, env := range envs {
		// apply remoteAgent filter
		if params.RemoteAgent != nil {
			hasAgent := agentEnvIds[env.Id]
			if *params.RemoteAgent != hasAgent {
				continue
			}
		}

		// apply hasActivity/idle filter
		hasActivity := false
		if activeEnvIds != nil {
			hasActivity = activeEnvIds[env.Id]

			// filter by hasActivity (wants active environments)
			if params.HasActivity != nil && *params.HasActivity && !hasActivity {
				continue
			}

			// filter by idle (wants inactive environments)
			if params.Idle != nil && *params.Idle && hasActivity {
				continue
			}
		}

		summary := &rest_model_zrok.EnvironmentSummary{
			EnvZID:      env.ZId,
			Description: env.Description,
			Host:        env.Host,
			Address:     env.Address,
			RemoteAgent: agentEnvIds != nil && agentEnvIds[env.Id],
			ShareCount:  int64(env.ShareCount),
			AccessCount: int64(env.AccessCount),
			Limited:     isLimited,
			CreatedAt:   env.CreatedAt.UnixMilli(),
			UpdatedAt:   env.UpdatedAt.UnixMilli(),
		}

		response.Environments = append(response.Environments, summary)
	}

	return metadata.NewListEnvironmentsOK().WithPayload(response)
}

// findEnvironmentsWithActivity queries InfluxDB to find which environments have metrics within the given duration
func findEnvironmentsWithActivity(envs []*store.EnvironmentWithCounts, duration time.Duration, influxCfg *metrics.InfluxConfig) (map[int]bool, error) {
	if len(envs) == 0 {
		return make(map[int]bool), nil
	}

	idb := influxdb2.NewClient(influxCfg.Url, influxCfg.Token)
	defer idb.Close()
	queryApi := idb.QueryAPI(influxCfg.Org)

	// build filter for environment IDs
	envFilter := "|> filter(fn: (r) =>"
	for i, env := range envs {
		if i > 0 {
			envFilter += " or"
		}
		envFilter += fmt.Sprintf(" r[\"envId\"] == \"%d\"", env.Id)
	}
	envFilter += ")"

	query := fmt.Sprintf("from(bucket: \"%v\")\n", influxCfg.Bucket) +
		fmt.Sprintf("|> range(start: -%v)\n", duration) +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
		envFilter + "\n" +
		"|> group(columns: [\"envId\"])\n" +
		"|> sum()"

	result, err := queryApi.Query(context.Background(), query)
	if err != nil {
		return nil, errors.Wrap(err, "error querying influxdb for environment activity")
	}

	activeEnvIds := make(map[int]bool)
	for result.Next() {
		envIdStr, ok := result.Record().ValueByKey("envId").(string)
		if !ok {
			continue
		}
		var envId int
		if _, err := fmt.Sscanf(envIdStr, "%d", &envId); err != nil {
			continue
		}
		// any non-zero value means there was activity
		if val, ok := result.Record().Value().(int64); ok && val > 0 {
			activeEnvIds[envId] = true
		}
	}

	if result.Err() != nil {
		return nil, errors.Wrap(result.Err(), "error reading influxdb query results")
	}

	return activeEnvIds, nil
}

// findEnvironmentsWithAgents checks which environments have agents enrolled
func findEnvironmentsWithAgents(envs []*store.EnvironmentWithCounts, trx *sqlx.Tx) (map[int]bool, error) {
	if len(envs) == 0 {
		return make(map[int]bool), nil
	}

	agentEnvIds := make(map[int]bool)

	for _, env := range envs {
		hasAgent, err := str.IsAgentEnrolledForEnvironment(env.Id, trx)
		if err != nil {
			return nil, errors.Wrapf(err, "error checking agent enrollment for environment %d", env.Id)
		}
		agentEnvIds[env.Id] = hasAgent
	}

	return agentEnvIds, nil
}
