package controller

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/controller/store"
)

func sparkDataForEnvironments(envs []*store.Environment) (rx, tx map[int][]int64, err error) {
	rx = make(map[int][]int64)
	tx = make(map[int][]int64)
	if len(envs) > 0 {
		qapi := idb.QueryAPI(cfg.Metrics.Influx.Org)

		envFilter := "|> filter(fn: (r) =>"
		for i, env := range envs {
			if i > 0 {
				envFilter += " or"
			}
			envFilter += fmt.Sprintf(" r[\"envId\"] == \"%d\"", env.Id)
		}
		envFilter += ")"
		query := fmt.Sprintf("from(bucket: \"%v\")\n", cfg.Metrics.Influx.Bucket) +
			"|> range(start: -5m)\n" +
			"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
			"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
			"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
			envFilter +
			"|> aggregateWindow(every: 10s, fn: sum, createEmpty: true)\n"

		result, err := qapi.Query(context.Background(), query)
		if err != nil {
			return nil, nil, err
		}

		for result.Next() {
			envId := result.Record().ValueByKey("envId").(int64)
			switch result.Record().Field() {
			case "rx":
				rxV := int64(0)
				if v, ok := result.Record().Value().(int64); ok {
					rxV = v
				}
				rxData := append(rx[int(envId)], rxV)
				rx[int(envId)] = rxData

			case "tx":
				txV := int64(0)
				if v, ok := result.Record().Value().(int64); ok {
					txV = v
				}
				txData := append(tx[int(envId)], txV)
				tx[int(envId)] = txData
			}
		}
	}
	return rx, tx, nil
}

func sparkDataForShares(shrs []*store.Share) (rx, tx map[string][]int64, err error) {
	rx = make(map[string][]int64)
	tx = make(map[string][]int64)
	if len(shrs) > 0 {
		qapi := idb.QueryAPI(cfg.Metrics.Influx.Org)

		shrFilter := "|> filter(fn: (r) =>"
		for i, shr := range shrs {
			if i > 0 {
				shrFilter += " or"
			}
			shrFilter += fmt.Sprintf(" r[\"share\"] == \"%v\"", shr.Token)
		}
		shrFilter += ")"
		query := fmt.Sprintf("from(bucket: \"%v\")\n", cfg.Metrics.Influx.Bucket) +
			"|> range(start: -5m)\n" +
			"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
			"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
			"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")\n" +
			shrFilter +
			"|> aggregateWindow(every: 10s, fn: sum, createEmpty: true)\n"

		result, err := qapi.Query(context.Background(), query)
		if err != nil {
			return nil, nil, err
		}

		for result.Next() {
			shrToken := result.Record().ValueByKey("share").(string)
			switch result.Record().Field() {
			case "rx":
				rxV := int64(0)
				if v, ok := result.Record().Value().(int64); ok {
					rxV = v
				}
				rxData := append(rx[shrToken], rxV)
				rx[shrToken] = rxData

			case "tx":
				txV := int64(0)
				if v, ok := result.Record().Value().(int64); ok {
					txV = v
				}
				txData := append(tx[shrToken], txV)
				tx[shrToken] = txData
			}
		}
	}
	return rx, tx, nil
}
