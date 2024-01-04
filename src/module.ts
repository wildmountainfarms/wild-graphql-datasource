import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { WildGraphQLMainQuery, WildGraphQLDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<DataSource, WildGraphQLMainQuery, WildGraphQLDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
