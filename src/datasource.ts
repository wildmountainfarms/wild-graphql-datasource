import {CoreApp, DataQueryRequest, DataQueryResponse, DataSourceInstanceSettings} from '@grafana/data';
import {DataSourceWithBackend, getTemplateSrv} from '@grafana/runtime';
import {Observable} from 'rxjs';

import {DEFAULT_QUERY, getQueryVariablesAsJson, WildGraphQLDataSourceOptions, WildGraphQLMainQuery} from './types';

export class DataSource extends DataSourceWithBackend<WildGraphQLMainQuery, WildGraphQLDataSourceOptions> {
  settings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>;

  constructor(instanceSettings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>) {
    super(instanceSettings);
    this.settings = instanceSettings;
  }

  getDefaultQuery(_: CoreApp): Partial<WildGraphQLMainQuery> {
    return DEFAULT_QUERY;
  }
  query(request: DataQueryRequest<WildGraphQLMainQuery>): Observable<DataQueryResponse> {

    // Everything you see going on here is to do variable substitution for the values of the provided variables.
    const templateSrv = getTemplateSrv();
    const newTargets: WildGraphQLMainQuery[] = request.targets.map((target) => {
      const variables = getQueryVariablesAsJson(target);
      const newVariables: any = { };
      for (const variableName in variables) {
        newVariables[variableName] = templateSrv.replace(variables[variableName]);
      }
      return {
        ...target,
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
}
