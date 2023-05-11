package controller

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/controller/store"
)

func sparkDataForShares(shrs []*store.Share) (rx, tx map[string][]int64, err error) {
	rx = make(map[string][]int64)
	tx = make(map[string][]int64)
	if len(shrs) > 0 {
		qapi := idb.QueryAPI(cfg.Metrics.Influx.Org)

		query := sparkFluxQuery(shrs, cfg.Metrics.Influx.Bucket)
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

func sparkFluxQuery(shrs []*store.Share, bucket string) string {
	shrFilter := "|> filter(fn: (r) =>"
	for i, shr := range shrs {
		if i > 0 {
			shrFilter += " or"
		}
		shrFilter += fmt.Sprintf(" r[\"share\"] == \"%v\"", shr.Token)
	}
	shrFilter += ")"
	query := fmt.Sprintf("from(bucket: \"%v\")\n", bucket) +
		"|> range(start: -5m)\n" +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")\n" +
		"|> filter(fn: (r) => r[\"_field\"] == \"rx\" or r[\"_field\"] == \"tx\")\n" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")" +
		shrFilter +
		"|> aggregateWindow(every: 10s, fn: sum, createEmpty: true)\n"
	return query
}
