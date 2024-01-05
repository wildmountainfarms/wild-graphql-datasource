import {CoreApp, DataSourceInstanceSettings} from '@grafana/data';
import {DataSourceWithBackend} from '@grafana/runtime';

import {DEFAULT_QUERY, WildGraphQLDataSourceOptions, WildGraphQLMainQuery} from './types';

export class DataSource extends DataSourceWithBackend<WildGraphQLMainQuery, WildGraphQLDataSourceOptions> {
  options: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>;

  constructor(instanceSettings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>) {
    super(instanceSettings);
    this.options = instanceSettings;
  }

  getDefaultQuery(_: CoreApp): Partial<WildGraphQLMainQuery> {
    return DEFAULT_QUERY;
  }
}
