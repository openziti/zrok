package controller

import (
	"context"
	"fmt"
	"github.com/openziti-test-kitchen/zrok/controller/store"
)

func sparkDataForServices(svcs []*store.Service) (map[string][]int64, error) {
	out := make(map[string][]int64)

	if len(svcs) > 0 {
		qapi := idb.QueryAPI(cfg.Influx.Org)

		result, err := qapi.Query(context.Background(), sparkFluxQuery(svcs))
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
			svcToken := result.Record().ValueByKey("service").(string)
			svcMetrics := out[svcToken]
			svcMetrics = append(svcMetrics, combinedRate)
			out[svcToken] = svcMetrics
		}
	}
	return out, nil
}

func sparkFluxQuery(svcs []*store.Service) string {
	svcFilter := "|> filter(fn: (r) =>"
	for i, svc := range svcs {
		if i > 0 {
			svcFilter += " or"
		}
		svcFilter += fmt.Sprintf(" r[\"service\"] == \"%v\"", svc.Token)
	}
	svcFilter += ")"
	query := "read = from(bucket: \"zrok\")" +
		"|> range(start: -5m)" +
		"|> filter(fn: (r) => r[\"_measurement\"] == \"xfer\")" +
		"|> filter(fn: (r) => r[\"_field\"] == \"bytesRead\" or r[\"_field\"] == \"bytesWritten\")" +
		"|> filter(fn: (r) => r[\"namespace\"] == \"frontend\")" +
		svcFilter +
		"|> aggregateWindow(every: 5s, fn: sum, createEmpty: true)\n" +
		"|> pivot(rowKey:[\"_time\"], columnKey: [\"_field\"], valueColumn: \"_value\")" +
		"|> yield(name: \"last\")"
	return query
}
