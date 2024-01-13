package util

import "github.com/grafana/grafana-plugin-sdk-go/backend"

func AutoPopulateVariables(query backend.DataQuery, variables *map[string]interface{}) {
	(*variables)["from"] = query.TimeRange.From.UnixMilli()
	(*variables)["to"] = query.TimeRange.To.UnixMilli()
	(*variables)["interval_ms"] = query.Interval.Milliseconds()
}

// When we do get around to supporting variable substitution on the backend, we should make it as similar to the frontend as possible:
//   https://github.com/grafana/grafana/blob/4b071f54529e24a2723eedf7ca4e7e989b3bd956/public/app/features/variables/utils.ts#L33
//   The reason we would want variable substitution at all is for annotation queries because you cannot transform the result of those queries in any way.
//   It might be possible to deal with that on the frontend, though.
