package plugin

import "github.com/grafana/grafana-plugin-sdk-go/backend"

func AutoPopulateVariables(query backend.DataQuery, variables *map[string]interface{}) {
	(*variables)["from"] = query.TimeRange.From.UnixMilli()
	(*variables)["to"] = query.TimeRange.To.UnixMilli()
	(*variables)["interval_ms"] = query.Interval.Milliseconds()
}
