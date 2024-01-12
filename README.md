# Wild GraphQL Datasource

**This is work in progress and is not in a working state.**

This is a Grafana datasource that aims to make requesting time series data via a GraphQL endpoint easy.
This datasource is similar to https://github.com/fifemon/graphql-datasource, but is not compatible.
This datasource tries to reimagine how GraphQL queries should be made from Grafana.

Requests are made in the backend. Results are consistent between queries and alerting.

## Uses for Timeseries data

### Using a field as the display name

If you have data that needs to be "grouped by" or "partitioned by", you first need to add "Partition by values"
as a transform and select `packet.identifier.representation` and `packet.identityInfo.displayName` as fields.
Once you do that, you can go into the "Standard options" and find "Display name" and set it to
`${__field.labels["packet.identityInfo.displayName"]}`.

References:

* [documentation for Display name under standard options](https://grafana.com/docs/grafana/latest/panels-visualizations/configure-standard-options/#display-name).
* [partition by values](https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform-data/transform-data/)
  * Note: Partition by values was [added in 9.3](https://grafana.com/docs/grafana/latest/whatsnew/whats-new-in-v9-3/#new-transformation-partition-by-values) ([blog](https://grafana.com/blog/2022/11/29/grafana-9.3-release/))

## Common errors

### Alerting Specific Errors

* `Failed to evaluate queries and expressions: input data must be a wide series but got type long (input refid)`
  * This error indicates that the query returns more fields than just the time and the datapoint.
  * For alerts, the response from the GraphQL query cannot contain more than the time and datapoint. At this time, you cannot use other attributes from the result to filter the data.

## To-Do

* Add metrics to backend component: https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-logs-metrics-traces-for-backend-plugins#implement-metrics-in-your-plugin
* Support returning logs data: https://grafana.com/developers/plugin-tools/tutorials/build-a-logs-data-source-plugin
  * We could just add `"logs": true` to `plugin.json`, however we need to support the renaming of fields because sometimes the `body` or `timestamp` fields will be nested
* Publish as a plugin
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/publish-a-plugin
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/sign-a-plugin#generate-an-access-policy-token
  * https://grafana.com/legal/plugins/
  * https://grafana.com/developers/plugin-tools/publish-a-plugin/provide-test-environment
