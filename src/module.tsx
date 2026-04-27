import React, { Suspense, lazy } from 'react';
import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import type { Props } from './components/QueryEditor';
import { WildGraphQLMainQuery, WildGraphQLDataSourceOptions } from './types';

const LazyQueryEditor = lazy(() => import('./components/QueryEditor'));

const QueryEditor = (props: Props) => {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <LazyQueryEditor {...props} />
    </Suspense>
  );
};

export const plugin = new DataSourcePlugin<DataSource, WildGraphQLMainQuery, WildGraphQLDataSourceOptions>(DataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
