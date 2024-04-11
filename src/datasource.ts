import {
  AnnotationSupport,
  CoreApp,
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
} from '@grafana/data';
import {DataSourceWithBackend, getTemplateSrv} from '@grafana/runtime';
import {Observable} from 'rxjs';

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
  query(request: DataQueryRequest<WildGraphQLAnyQuery>): Observable<DataQueryResponse> {

    // Everything you see going on here is to do variable substitution for the values of the provided variables.
    const templateSrv = getTemplateSrv();
    const newTargets: WildGraphQLAnyQuery[] = request.targets.map((target) => {
      const variables = getQueryVariablesAsJson(target);
      const interpolatedVariables = interpolateVariables(variables, templateSrv, request.scopedVars);
      const interpolatedVariablesWithFullInterpolationObject = target.variablesWithFullInterpolation === undefined
        ? {}
        : JSON.parse( // TODO JSON.parse may throw a SyntaxError. For now, we're OK with that, but we may consider how to make this so it doesn't prevent ALL queries from being executed
          templateSrv.replace(target.variablesWithFullInterpolation, request.scopedVars)
        );
      const newVariables = {
        ...interpolatedVariables,
        ...interpolatedVariablesWithFullInterpolationObject
      }
      return {
        ...target,
        variablesWithFullInterpolation: undefined, // the backend does not care about this value, so don't pass it
        variables: newVariables,
      }
    })
    const newRequest = {
      ...request,
      targets: newTargets
    };

    // we aren't really supposed to change this method, but we do it anyway :)
    return super.query(newRequest);
  }
  // metricFindQuery(query: any, options?: any): Promise<MetricFindValue[]> {
  //   // https://grafana.com/developers/plugin-tools/create-a-plugin/extend-a-plugin/add-support-for-variables
  //   // Note that in the future, it looks like metricFindQuery will be deprecated in favor of variable support, similar in style to annotation support
  //   return super.metricFindQuery(query, options);
  //   // Once we add this, use this query on the provisioned dashboard: `query {film(id: "ZmlsbXM6MQ==") {openingCrawl}}`
  // }
}
