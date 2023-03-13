package controller

import (
	"context"
	"fmt"
	"github.com/openziti/zrok/controller/store"
)

func sparkDataForShares(shrs []*store.Share) (map[string][]int64, error) {
	out := make(map[string][]int64)

	if len(shrs) > 0 {
		qapi := idb.QueryAPI(cfg.Metrics.Influx.Org)

		result, err := qapi.Query(context.Background(), sparkFluxQuery(shrs))
		if err != nil {
			return nil, err
		}

		for result.Next() {
			combinedRate := int64(0)
			readRate := result.Record().ValueByKey("bytesRead")
			if readRate != nil {
				combinedRate += readRate.(int64)
			}
			writeRate := result.Record().ValueByKey("bytesWritten")
			if writeRate != nil {
				combinedRate += writeRate.(int64)
			}
			shrToken := result.Record().ValueByKey("share").(string)
			shrMetrics := out[shrToken]
			shrMetrics = append(shrMetrics, combinedRate)
			out[shrToken] = shrMetrics
		}
	}
	return out, nil
}

func sparkFluxQuery(shrs []*store.Share) string {
	shrFilter := "|> filter(fn: (r) =>"
	for i, shr := range shrs {
		if i > 0 {
			shrFilter += " or"
		}
		shrFilter += fmt.Sprintf(" r[\"share\"] == \"%v\"", shr.Token)
	}
	shrFilter += ")"
	query := "read = from(bucket: \"zrok\")" +
		"|> range(start: -5m)" +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesRead\" or r[\"_field\"] == \"bytesWritten\")" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"backend\")" +
		shrFilter +
		"|> aggregateWindow(every: 5s, fn: sum, createEmpty: true)\n" +
		"|> pivot(rowKey:[\"_time\"], columnKey: [\"_field\"], valueColumn: \"_value\")" +
		"|> yield(name: \"last\")"
	return query
}
