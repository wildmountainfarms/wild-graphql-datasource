import React, { ChangeEvent } from 'react';
import { InlineField, Input } from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { WildGraphQLDataSourceOptions, WildGraphQLMainQuery } from '../types';

type Props = QueryEditorProps<DataSource, WildGraphQLMainQuery, WildGraphQLDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, queryText: event.target.value });
  };

  const { queryText } = query;

  return (
    <div className="gf-form">
      <InlineField label="Query Text" labelWidth={16} tooltip="The GraphQL query">
        <Input onChange={onQueryTextChange} value={queryText} />
      </InlineField>
    </div>
  );
}
