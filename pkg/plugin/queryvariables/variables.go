package queryvariables

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"reflect"
)

func AutoPopulateVariables(query backend.DataQuery, variables map[string]interface{}) {

	// Although the frontend has access to global variable substitution (https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#global-variables)
	//   the backend does not.
	//   Because of this, it's beneficial to encourage users to write queries that rely as little on the frontend as possible.
	//   This allows alerting queries to be supported.
	//   These variable names that we are "hard coding" should be as similar to possible as those global variables that are available
	//   Forum post here: https://community.grafana.com/t/how-to-use-template-variables-in-your-data-source/63250#backend-data-sources-3
	//   More information here: https://grafana.com/docs/grafana/latest/dashboards/variables/

	variables["from"] = query.TimeRange.From.UnixMilli()
	variables["to"] = query.TimeRange.To.UnixMilli()
	variables["interval_ms"] = query.Interval.Milliseconds()
	variables["maxDataPoints"] = query.MaxDataPoints
	variables["refId"] = query.RefID
}

// When we do get around to supporting variable substitution on the backend, we should make it as similar to the frontend as possible:
//   https://github.com/grafana/grafana/blob/4b071f54529e24a2723eedf7ca4e7e989b3bd956/public/app/features/variables/utils.ts#L33
//   The reason we would want variable substitution at all is for annotation queries because you cannot transform the result of those queries in any way.
//   It might be possible to deal with that on the frontend, though.

func ParseVariables(query backend.DataQuery, rawVariables interface{}) (map[string]interface{}, bool) {
	var noErrors = true
	variables := map[string]interface{}{}

	// call AutoPopulateVariables first so that users can override variables if they see fit
	// NOTE: When we add support for secure variables, overriding should NOT be supported
	AutoPopulateVariables(query, variables)

	switch typedRawVariables := rawVariables.(type) {
	case string:
		// This case happens when the frontend is not involved at all. This is likely an alert.
		// Remember that these variables are not interpolated

		err := json.Unmarshal([]byte(typedRawVariables), &variables)
		if err != nil {
			noErrors = false
			log.DefaultLogger.Error("Got error while parsing variables!", "typedRawVariables", typedRawVariables, "err", err)

			// continue executing query without interpolated variables
			// TODO consider if we want a flag in the options to prevent the query from continuing further in the case of an error
		}
	case map[string]interface{}:
		// This case happens when the frontend is able to interpolate the variables before passing them to us
		//   or happens when someone has directly configured the variables option in the JSON itself
		//   and storing it as an object rather than a string like the QueryEditor does.
		//   If this is the ladder case, variables have not been interpolated
		for key, value := range typedRawVariables {
			variables[key] = value
		}
	case nil:
		// do nothing
	default:
		noErrors = false
		log.DefaultLogger.Error(fmt.Sprintf("Unable to parse variables for ref ID: %s. Type is %v", query.RefID, reflect.TypeOf(rawVariables)))
	}

	return variables, noErrors
}
