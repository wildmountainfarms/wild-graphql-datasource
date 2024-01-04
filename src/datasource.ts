import { DataSourceInstanceSettings, CoreApp } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';

import { WildGraphQLMainQuery, WildGraphQLDataSourceOptions, DEFAULT_QUERY } from './types';

export class DataSource extends DataSourceWithBackend<WildGraphQLMainQuery, WildGraphQLDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<WildGraphQLDataSourceOptions>) {
    super(instanceSettings);
  }

  getDefaultQuery(_: CoreApp): Partial<WildGraphQLMainQuery> {
    return DEFAULT_QUERY;
  }
}
