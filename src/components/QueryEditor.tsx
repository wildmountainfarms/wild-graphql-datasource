import React, {ChangeEvent, useMemo} from 'react';
import {InlineField, Input} from '@grafana/ui';
import {QueryEditorProps} from '@grafana/data';
import {DataSource} from '../datasource';
import {WildGraphQLDataSourceOptions, WildGraphQLMainQuery} from '../types';
import {GraphiQLInterface} from 'graphiql';
import {
  EditorContextProvider,
  ExplorerContextProvider,
  PluginContextProvider,
  SchemaContextProvider,
  ExecutionContextProvider,
  // StorageContextProvider,
  // HistoryContextProvider
} from '@graphiql/react';
import {Fetcher} from '@graphiql/toolkit';
import {FetcherOpts, FetcherParams} from "@graphiql/toolkit/src/create-fetcher/types";
import {getBackendSrv} from "@grafana/runtime";
import {firstValueFrom} from 'rxjs';

import 'graphiql/graphiql.css';

// import '@graphiql/react/dist/style.css';
// import '@graphiql/react/font/roboto.css';
// import '@graphiql/react/font/fira-code.css';



type Props = QueryEditorProps<DataSource, WildGraphQLMainQuery, WildGraphQLDataSourceOptions>;

/**
 * This fetcher is designed to be used only for fetching the schema of a GraphQL endpoint.
 * This uses [getBackendSrv] to use Grafana's default backend HTTP proxy.
 * This means that we make requests to the GraphQL endpoint in two different ways, this being the less common and less robust way.
 * This is less robust because DataSourceHttpSettings defines many different options, and we don't actually respect all of them here.
 */
function createFetcher(url: string, withCredentials: boolean, basicAuth?: string): Fetcher  {
  const headers: Record<string, any> = {
    Accept: 'application/json',
    'Content-Type': 'application/json',
  };
  if (withCredentials) { // TODO is this how withCredentials is supposed to be used?
    headers['Authorization'] = basicAuth;
  }
  const backendSrv = getBackendSrv();
  return async (graphQLParams: FetcherParams, opts?: FetcherOpts) => {
    const observable = backendSrv.fetch({
      url,
      headers,
      method: "POST",
      data: graphQLParams,
      responseType: "json",
      // TODO consider other options available here
    });
    // TODO handle error cases here
    const response = await firstValueFrom(observable);
    return response.data;
    // const data = await fetch(url || '', {
    //   method: 'POST',
    //   headers: headers,
    //   body: JSON.stringify(graphQLParams),
    //   credentials: 'same-origin',
    // });
    // return data.json().catch(() => data.text());
  };
}

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {


  const fetcher = useMemo(() => {
    return createFetcher(
      datasource.options.url!,
      datasource.options.withCredentials ?? false,
      datasource.options.basicAuth
    );
  }, [datasource.options.url, datasource.options.withCredentials, datasource.options.basicAuth]);

  const onOperationNameChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, operationName: event.target.value || undefined });
  };

  return (
    <>
      <h3 className="page-heading">Query</h3>
      <div className="gf-form-group">
        <div className="gf-form">
          {/*<InlineFormLabel width={13}>Query</InlineFormLabel>*/}
          {/*<GraphiQL*/}
          {/*  fetcher={fetcher}*/}
          {/*  defaultQuery={query.queryText}*/}
          {/*  onEditQuery={(value) => {*/}
          {/*    console.log("Edited query");*/}
          {/*    console.log(value);*/}
          {/*    onChange({...query, queryText: value});*/}
          {/*  }}*/}
          {/*  isHeadersEditorEnabled={false}*/}
          {/*  showPersistHeadersSettings={false}*/}
          {/*  storage={DummyStorage}*/}
          {/*  shouldPersistHeaders={false}*/}
          {/*  plugins={}*/}
          {/*/>*/}

          {/*By not providing storage, history contexts, they won't be used*/}
          {/*<StorageContextProvider storage={DummyStorage}>*/}
          {/*  <HistoryContextProvider maxHistoryLength={0}>*/}
          <EditorContextProvider
            defaultQuery={query.queryText}
          >
            <SchemaContextProvider fetcher={fetcher}>
              <ExecutionContextProvider
                fetcher={fetcher}
                // TODO consider passing operationName here
              >
                <ExplorerContextProvider> {/*Explorer context needed for documentation*/}
                  <PluginContextProvider>
                    <GraphiQLInterface
                      showPersistHeadersSettings={false}
                      // TODO add disableTabs={true} when release supports https://github.com/graphql/graphiql/pull/3408
                      isHeadersEditorEnabled={false}
                      onEditQuery={(value) => {
                        console.log("Edited query");
                        console.log(value);
                        onChange({...query, queryText: value});
                      }}
                    />
                  </PluginContextProvider>
                </ExplorerContextProvider>
              </ExecutionContextProvider>
            </SchemaContextProvider>
          </EditorContextProvider>

        </div>
        <div className="gf-form-inline">
          <InlineField label="Operation Name" labelWidth={32}
                       tooltip="The operationName passed to the GraphQL endpoint. This can be left blank unless you specify multiple queries.">
            <Input onChange={onOperationNameChange} value={query.operationName ?? ''}/>
          </InlineField>
        </div>
      </div>
    </>
  );
}
