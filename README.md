# Wild GraphQL Data Source

[![](https://img.shields.io/github/stars/wildmountainfarms/wild-graphql-datasource.svg?style=social)](https://github.com/wildmountainfarms/wild-graphql-datasource)
[![](https://img.shields.io/github/v/release/wildmountainfarms/wild-graphql-datasource.svg)](https://github.com/wildmountainfarms/wild-graphql-datasource/releases)

This is a Grafana data source that aims to make requesting time series data via a GraphQL endpoint easy.
The query editor uses [GraphiQL](https://github.com/graphql/graphiql) to provide an intuitive editor with autocompletion.
Requests are made in the backend allowing support for alerting.

Please report issues on our issues page: [wild-graphql-datasource/issues](https://github.com/wildmountainfarms/wild-graphql-datasource/issues)

Contents

* [Features](#features)
* [Query Editor](#query-editor)
  * [Query](#query)
  * [Variables](#variables)
* [Parsing Options](#parsing-options)
  * [Data path and time path](#data-path-and-time-path)
  * [Labels](#labels)
  * [The use of multiple parsing options](#the-use-of-multiple-parsing-options)
* [FAQ](#faq)
* [Common Errors](#common-errors)
* [Known Issues](#known-issues)


## Features

* Complex GraphQL responses can be turned into timeseries data, or a simple table
* Includes [GraphiQL](https://github.com/graphql/graphiql) query editor. Autocompletion and documentation for the GraphQL schema available inside Grafana!
  * Documentation explorer can be opened from the query editor
  * Prettify the query with the click of a button
* `from` and `to` variables are given to the query via [native GraphQL variables](https://graphql.org/learn/queries/#variables)
* Variables section of the query editor supports interpolation of string values using [Grafana variables](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/). (\*not supported in alerting or other backend-only queries)
* Multiple parsing options are supported allowing for a single GraphQL query to return many different data point with different formats.
  * Each parsing option has its own labels, which can be populated by a field in the response. These labels are used to group the response into different data frames.
  * Labels can be used to change the display name by using `${__field.labels["displayName"]}` under Standard options > Display name.
* This is a backend plugin, so alerting is supported
* [Annotation support](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/annotate-visualizations/)

---

## Query Editor

### Query

Your GraphQL query goes inside the upper half of the GraphiQL editor pane

### Variables

Variables may be edited within the GraphiQL editor window.
These variables are specified using JSON and work like variables in any regular GraphQL query.
This section is optional, but if you wish to provide constant variables for your query or 
[simplify the specification of input types](https://graphql.org/graphql-js/mutations-and-input-types/), you may do so.

#### Provided Variables

Certain variables are provided to every query. These include:

| Variable        | Type   | Description                                                                    | Grafana counterpart                                                                                               | Execute Button Support |
|-----------------|--------|--------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|------------------------|
| `from`          | Number | Epoch milliseconds of the "from" time                                          | [$__from](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to)   | Yes                    |
| `to`            | Number | Epoch milliseconds of the "to" time                                            | [$__to](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to)     | Yes                    |
| `interval_ms`   | Number | The suggested duration between time points in a time series query              | [$__interval_ms](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__interval) | No                     |
| `maxDataPoints` | Number | Maximum number of data points that should be returned from a time series query | N/A                                                                                                               | No                     |
| `refId`         | String | Unique identifier of the query, set by the frontend call                       | N/A                                                                                                               | No                     |

An example usage is shown in the most basic query:

```graphql
query ($from: Long!, $to: Long!) {
  queryStatus(from: $from, to: $to) {
    # ...
  }
}
```

In the above example, the query asks for two Longs, `$from` and `$to`.
The value is provided by the provided variables as seen in the above table.
Notice that while `interval_ms` is provided, we do not use it or define it anywhere in our query.
One thing to keep in mind for your own queries is the type accepted by the GraphQL server for a given variable.
In the case of that specific schema, the type of a `Long` is allowed to be a number or a string.
If your schema does not have a `Long` type, you may consider changing the declaration of the query to use `Float` or `String`,
which are both types built into GraphQL.

If you run into any issues using the `from` and `to` variables for your specific schema, 
please [raise and issue](https://github.com/wildmountainfarms/wild-graphql-datasource/issues).
Alternatively, if you are making a query on the frontend (any query besides alerting queries),
you may consider reading the next section to define your own from and to variables and assigning their values to
[a global variable](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to).


#### Grafana variable interpolation

The variables section is the most useful for variable interpolation.
Any value inside a string, whether that string is nested inside an object, or a top-most value of the variables object, can be interpolated.
Please note that interpolation does not work for alerting queries.
You may use any configuration of [variables](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/) you see fit.
An example is this:

```graphql
query ($sourceId: String!, $from: Long!, $to: Long!) {
  queryStatus(sourceId: $sourceId, from: $from, to: $to) {
    # ...
  }
}
```

```json
{
  "sourceId": "$sourceId"
}
```

Here, `$sourceId` inside of the variables section will be interpolated with a value defined in your Grafana dashboard.
`$sourceId` inside of the GraphQL query pane is a regular [variable](https://graphql.org/learn/queries/#variables) that is passed to the query.

NOTE: Interpolating the entirety of the JSON text is not supported at this time.
This means that interpolated variables cannot be passed as numbers and interpolated variables cannot define complex JSON objects.
One of the pros of this is simplicity, with the advantage of not having to worry about escaping your strings.

REMEMBER: Variable interpolation does not work for alerting queries, or any query that is executed without the frontend component.

#### Advanced Variable Interpolation

If you need to incorporate numeric variables in the variables passed to your query, you cannot use the default variables editor, you must instead use Advanced Variables JSON.
Click the checkbox to define it. Once you do, you can write JSON that is not valid until AFTER it is interpolated. For instance:

```json
{
  "age": $age
}
```

The above example is valid assuming that `$age` evaluates to a number that causes the resulting JSON to be valid.
You can use both advanced variables JSON and the regular variables as described above. Just realize that advanced variables JSON will take precedence for any variables defined twice.

WARNING: Only use advanced variables JSON if it is required to. If any part of the resulting JSON is invalid, no part of the advanced variable interpolation will be passed to the query.

### Documentation Explorer

[GraphiQL](https://github.com/graphql/graphiql) provides a documentation explorer and can be opened on the left side of the GraphiQL editor
by clicking the button in the upper left of the GraphiQL editor.

### GraphiQL Query Execution

The right side of the GraphiQL editor is reserved for the results of the query.
This is useful only for debugging to see what the resulting JSON is of the query.

To run a query and see its raw JSON, press the "Execute query" button, which is in the top center of the GraphiQL editor.
The results you see in here are separate from allowing Grafana to query the data source.

### Operation Name

The operation name is displayed right below the GraphiQL editor.
The operation name should be automatically determined by Wild GraphQL Data Source.
If you define multiple queries in the query pane, you may have to manually specify this.
This is also automatically updated after running a query using the "Execute query" button.

## Parsing Options

The query editor allows you to configure one or many parsing options.
Each additional parsing option results in at least one additional data frame,
which can be seen in "Table View" of the Edit panel page on in the [query inspector](https://grafana.com/docs/grafana/latest/panels-visualizations/panel-inspector/).

### Data path and time path

The data path is a dot-delimited path to the array of data, or to the data object in the response JSON.

The time path is a dot-delimited path to the time field, or blank if there is no time field.

Take this query for example:

```graphql
query ($from: Long!, $to: Long!) {
  temperatures(from: $from, to: $to) {
    epochTimeMillis
    temperatureCelsius
  }
}
```

* Data path: `temperatures`
* Time path: `epochTimeMillis`

Notice that time path is relative to the data path.

### Labels

Each query option may specify labels that will be present in the resulting dataframe.
You may create a new label by typing in the "Label to add" text box and pressing enter.
This will add a label with the given name to each parsing option.

#### Field Labels

The value of a field label is the dot-delimited path to the desired field present in the response JSON,
similar to how the "Time Path" works.

The use of a field label will partition the data by that field's value, which results in multiple data frames,
which in turn results in multiple keys on the legend of your graph.

#### Constant Labels

A constant label defines a label for the given parsing option.
A constant label can be useful to identify the different parsing options,
or to provide a particular display name for a single parsing option.

#### Using a label as a display name

If you want to partition the data by a certain field, you should use a field label as described above.
With that field label created, there's a good chance you want to use that same field as the display name, so the legend of your graph shows a readable display name.
On the right hand side of the panel editor, navigate to [Standard Options](https://grafana.com/docs/grafana/latest/panels-visualizations/configure-standard-options/) > [Display Name](https://grafana.com/docs/grafana/latest/panels-visualizations/configure-standard-options/#display-name).
Now, change the display name to `${__field.labels["displayName"]}`, where `displayName` is the name of your label.

Alternatively, assuming your label's name is a [valid identifier](https://developer.mozilla.org/en-US/docs/Glossary/Identifier),
you may instead set the [Display Name](https://grafana.com/docs/grafana/latest/panels-visualizations/configure-standard-options/#display-name)
to `${__field.labels.displayName}`.

### The use of multiple parsing options

Multiple parsing options are a good alternative to multiple queries in a single panel.
Since a single GraphQL query can return lots of data, sometimes you may need to parse individual parts of it.
For instance, here's a simple example with 2 parsing options:

```graphql
query ($from: Long!, $to: Long!) {
  temperature(from: $from, to: $to) {
    epochTimeMillis
    sensorName
    temperatureCelsius
  }
  deviceCpuTemperature(from: $from, to: $to) {
    dateMillis
    temperature
  }
}
```

* Parsing option 1
  * Data path: `temperatures`
  * Time path: `epochTimeMillis`
  * Labels
    *  "displayName": (Field) `sensorName`
* Parsing option 2
  * Data path: `deviceCpuTemperatures`
  * Time path: `dateMillis`
  * Labels
    * "displayName": (Constant) `CPU`

What could have taken two queries, you now have done in a single query!

### Using Grafana Transformations

It's worth documenting that it's entirely possible to not use the labels feature of Wild GraphQL Data Source,
and use Grafana transformations for some of the same functionality.

If you have data that needs to be "grouped by" or "partitioned by", you first need to add "Partition by values"
as a transform and select `your.field.for.displayName`.
Once you do that, you can go into the "Standard options" and find "Display name" and set it to
`${__field.labels["your.field.for.displayName"]}`.

References:

* [documentation for Display name under standard options](https://grafana.com/docs/grafana/latest/panels-visualizations/configure-standard-options/#display-name).
* [partition by values](https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform-data/transform-data/)
  * Note: Partition by values was [added in 9.3](https://grafana.com/docs/grafana/latest/whatsnew/whats-new-in-v9-3/#new-transformation-partition-by-values) ([blog](https://grafana.com/blog/2022/11/29/grafana-9.3-release/))

Remember that Grafana transformations are not the preferred way of doing this, and you should prefer to use the label functionality provided.

---

## FAQ

* Can I use variable interpolation within the GraphQL query itself?
  * No, but you may use variable interpolation inside the string values of the variables passed to the query
* Is this a drop-in replacement for [fifemon-graphql-datasource](https://grafana.com/grafana/plugins/fifemon-graphql-datasource/)?
  * No, but both data sources have similar goals and can be ported between with little effort.

## Common errors

This section documents errors that may be common

### Alerting Specific Errors

* `Failed to evaluate queries and expressions: input data must be a wide series but got type long (input refid)`
  * This error indicates that the query returns more fields than just the time and the datapoint.
  * For alerts, the response from the GraphQL query cannot contain more than the time and datapoint. At this time, you cannot use other attributes from the result to filter the data.
* `Failed to evaluate queries and expressions: failed to execute conditions: input data must be a wide series but got type not (input refid)`
  * This may occur if you don't have any numeric data in your response (https://github.com/grafana/grafana/issues/46429)
  * This error also occurs when you include labels and your response includes multiple data frames.
    * To fix this, don't use labels in alerting queries. Alerting queries support the long data frame format, so it will automatically assume non-numeric fields are labels.
    * This has the drawback that if your query returns data across a period of time, you cannot easily partition data by fields AND choose to only use the most recent data.

## Known Issues

* Alerting queries and annotation queries can only use fields provided in the response data.
  * We plan to mitigate this in the future by allowing custom fields to be added to the response
