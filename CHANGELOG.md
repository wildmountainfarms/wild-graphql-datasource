# Changelog

## 1.3.0

Merged [#8](https://github.com/wildmountainfarms/wild-graphql-datasource/pull/8)
which passes a request's HTTP headers to the GraphQL server.
This change should respect the checkboxes under the "Auth" section of the data source's configuration
so that OAuth and cookie headers are only sent if you toggle their setting.

Additionally, these headers are now always passed to the GraphQL server for most queries:

* `X-Datasource-Uid`
* `X-Grafana-Org-Id`
* `X-Panel-Id`
* `X-Dashboard-Uid`

## 1.2.1

Updated LICENSE link in README.

## 1.2.0

Merged [#5](https://github.com/wildmountainfarms/wild-graphql-datasource/pull/5)
which updates the internal libraries for better HTTP header support that can be configured within the data source itself.
Additionally, all GraphQL requests will include a `Accept: application/json` header in every request.
This additional header matches [the GraphQL over HTTP spec](https://graphql.github.io/graphql-over-http/draft/#sec-Accept) for better compatibility with GraphQL servers.

## 1.1.1

No new features or fixes.

Stability
* First release that is officially signed

## 1.1.0

New Features
* Added "Advanced Variables JSON" to define a variables JSON object that has interpolation performed on the JSON string itself, rather than strings within the JSON
  * This is added functionality and is not meant to replace or change the existing functionality. Using the variables configuration from the GraphiQL editor and advanced variables JSON at the same time is supported.

Stability
* Resources are properly closed
* Panics are no longer used and better logging is done in the backend
* Unnecessary `console.log()` calls moved to existing `console.error()` calls


## 1.0.1

Fixes
* Numeric values are now exported as `float64` values, which allows alerting queries to work correctly
* An undefined variables object used to cause an error to be logged. Null/undefined `variables` field defined via provisioned queries now cause no error message to be logged.


## 1.0.0

Initial release.
