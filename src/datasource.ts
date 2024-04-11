import {AdHocVariableFilter, AnnotationSupport, CoreApp, DataSourceInstanceSettings, ScopedVars,} from '@grafana/data';
import {DataSourceWithBackend, getTemplateSrv} from '@grafana/runtime';

import {
  DEFAULT_ALERTING_QUERY,
  DEFAULT_ANNOTATION_QUERY,
  DEFAULT_QUERY,
  getQueryVariablesAsJson,
  WildGraphQLAnnotationQuery,
  WildGraphQLAnyQuery,
  WildGraphQLDataSourceOptions,
} from './types';
import {interpolateVariables} from "./variables";

export class DataSource extends DataSourceWithBackend<WildGraphQLAnyQuery, WildGraphQLDataSourceOptions> {
  settings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>;
  annotations: AnnotationSupport<WildGraphQLAnnotationQuery>;

  constructor(instanceSettings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>) {
    super(instanceSettings);
    this.settings = instanceSettings;
    this.annotations = {
      // TODO annotation support is very minimal right now
      // It works perfectly fine, however we need a way to add additional fields that can have template interpolation
      //   to create informational labels for the annotation
      getDefaultQuery(): Partial<WildGraphQLAnnotationQuery> {
        return DEFAULT_ANNOTATION_QUERY;
      }
    };
  }

  getDefaultQuery(app: CoreApp): Partial<WildGraphQLAnyQuery> {
    if (app === CoreApp.CloudAlerting || app === CoreApp.UnifiedAlerting) {
      // we have a different default query for alerts because alerts only support returning time and value columns.
      //   Additional columns in the data frame will return in an "input data must be a wide series" error.
      return DEFAULT_ALERTING_QUERY;
    }
    return DEFAULT_QUERY;
  }
  applyTemplateVariables(query: WildGraphQLAnyQuery, scopedVars: ScopedVars, filters?: AdHocVariableFilter[]): WildGraphQLAnyQuery {
    const templateSrv = getTemplateSrv();
    const variables = getQueryVariablesAsJson(query);
    const interpolatedVariables = interpolateVariables(variables, templateSrv, scopedVars);
    let interpolatedVariablesWithFullInterpolation: Record<string, any> = {};
    if (query.variablesWithFullInterpolation !== undefined) {
      // variablesWithFullInterpolation are inherently unsafe and should only be used when necessary.
      //   We have to check to make sure parsing it does not result in an error.
      //   We can only parse it after calling templateSrv.replace()
      const interpolatedString = templateSrv.replace(query.variablesWithFullInterpolation, scopedVars);
      try {
        interpolatedVariablesWithFullInterpolation = JSON.parse(interpolatedString);
      } catch (e) {
        if (e instanceof SyntaxError) {
          console.error("Error parsing variablesWithFullInterpolation for refId: " + query.refId + ". argument to JSON.parse(), error:", interpolatedString, e);
          // We don't have a good way of indicating to the caller than an error has occurred,
          //   so we instead continue without merging the variablesWithFullInterpolation
        } else {
          throw e;
        }
      }
    }
    const newVariables = {
      ...interpolatedVariables,
      ...interpolatedVariablesWithFullInterpolation
    };
    return {
      ...query,
      variablesWithFullInterpolation: undefined, // the backend does not care about this value, so don't pass it
      variables: newVariables,
    };
  }
  // metricFindQuery(query: any, options?: any): Promise<MetricFindValue[]> {
  //   // https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-support-for-variables
  //   // Note that in the future, it looks like metricFindQuery will be deprecated in favor of variable support, similar in style to annotation support
  //   return super.metricFindQuery(query, options);
  //   // Once we add this, use this query on the provisioned dashboard: `query {film(id: "ZmlsbXM6MQ==") {openingCrawl}}`
  // }
}
