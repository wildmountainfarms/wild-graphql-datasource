import React, {ChangeEvent, useEffect, useMemo} from 'react';
import {Button, IconButton, InlineField, Input} from '@grafana/ui';
import {CoreApp, QueryEditorProps} from '@grafana/data';
import {DataSource} from '../datasource';
import {
  getQueryVariablesAsJsonString,
  ParsingOption,
  WildGraphQLAnyQuery,
  WildGraphQLDataSourceOptions
} from '../types';
import {GraphiQLInterface} from 'graphiql';
import {
  EditorContextProvider,
  ExecutionContextProvider,
  ExplorerContextProvider,
  PluginContextProvider,
  SchemaContextProvider,
  useEditorContext
} from '@graphiql/react';
import {Fetcher} from '@graphiql/toolkit';
import {FetcherOpts, FetcherParams} from "@graphiql/toolkit/src/create-fetcher/types";
import {getBackendSrv, getTemplateSrv} from "@grafana/runtime";
import {firstValueFrom} from 'rxjs';

import 'graphiql/graphiql.css';
import './modify_graphiql.css'
import {ExecutionResult} from "graphql-ws";
import {getInterpolatedAutoPopulatedVariables, interpolateVariables} from "../variables";


type Props = QueryEditorProps<DataSource, WildGraphQLAnyQuery, WildGraphQLDataSourceOptions>;

const LABEL_WIDTH = 24;
const INPUT_WIDTH = 48;

/**
 * This fetcher is designed to be used only for fetching the schema of a GraphQL endpoint.
 * This uses {@link getBackendSrv} to use Grafana's default backend HTTP proxy.
 * This means that we make requests to the GraphQL endpoint in two different ways, this being the less common and less robust way.
 * This is less robust because DataSourceHttpSettings defines many different options, and we don't actually respect all of them here.
 *
 * This fetcher also automatically performs variable templating using {@link getTemplateSrv}.
 * This templating is only applied to the variables themselves, not the queryText.
 * This is useful for when pressing the run button on the query editor itself, which (just like the schema)
 * is not sent through the more robust backend logic.
 * This is consistent with how the query should be altered on the frontend before sending it to the backend.
 * One key difference here is that it is expected that all variables populated automatically by the backend
 * are also automatically populated by this method, using
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
  // NOTE: getTemplateSrv() is something that is only updated after a query is performed on a Grafana dashboard.
  //   If you navigate straight to "Alert rules", for example, getTemplateSrv() will not be able to replace $__to and $__from variables.
  //   This has the implication that the "Execute query" button performs a query with "to" and "from" variables that are unlike what is actually configured.
  const templateSrv = getTemplateSrv();
  return async (graphQLParams: FetcherParams, opts?: FetcherOpts) => {
    const variables = {
      ...getInterpolatedAutoPopulatedVariables(templateSrv),
      ...interpolateVariables(graphQLParams.variables, templateSrv), // remember one of the downsides here is that we cannot pass scopedVars here because we don't have access to it
    };
    const query = {
      ...graphQLParams,
      variables: variables
    };
    console.log(query);
    const observable = backendSrv.fetch({
      url,
      headers,
      method: "POST",
      data: query,
      responseType: "json",
      // NOTE: Other options may be necessary here, but at the time of writing I have not tested the different scenarios that might warrant a need to alter these parameters
    });
    // awaiting the observable may throw an exception, and that's OK, we can let that propagate up
    const response = await firstValueFrom(observable);
    return response.data as ExecutionResult;
  };
}

export function QueryEditor(props: Props) {
  const { query, datasource } = props;
  const isAlerting = props.app === CoreApp.CloudAlerting || props.app === CoreApp.UnifiedAlerting;

  const fetcher = useMemo(() => {
    return createFetcher(
      datasource.settings.url!,
      datasource.settings.withCredentials ?? false,
      datasource.settings.basicAuth
    );
  }, [datasource.settings.url, datasource.settings.withCredentials, datasource.settings.basicAuth]);

  return (
    <>
      {/*By not providing storage, history contexts, they won't be used*/}
      {/*<StorageContextProvider storage={DummyStorage}>*/}
      {/*  <HistoryContextProvider maxHistoryLength={0}>*/}
      <EditorContextProvider
        // defaultQuery is the query that is used for new tabs, but we already define the open tabs here
        defaultTabs={[{
          query: query.queryText,
          // NOTE: For some reason if you specify variable here, it just doesn't work...
        }]}
        variables={getQueryVariablesAsJsonString(query)}
        // we don't need to pass onEditOperationName here because we have a callback that handles it ourselves
      >
        <SchemaContextProvider fetcher={fetcher}>
          <ExecutionContextProvider
            fetcher={fetcher}
            // NOTE: We don't pass the operationName here because when the user presses the run button,
            //   we want them to always have to choose which operation they want
          >
            <ExplorerContextProvider> {/*Explorer context needed for documentation*/}
              <PluginContextProvider>
                {/*We need to hide the execute button and response window during alerting because the to and from variables are not populated correctly*/}
                <div className={isAlerting ? "hide-execute-button" : ""}>
                  <InnerQueryEditor
                    {...props}
                  />
                </div>
              </PluginContextProvider>
            </ExplorerContextProvider>
          </ExecutionContextProvider>
        </SchemaContextProvider>
      </EditorContextProvider>
    </>
  );
}

function InnerQueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const editorContext = useEditorContext();
  const onOperationNameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const newOperationName = event.target.value || undefined;
    const queryEditor = editorContext?.queryEditor;
    if (queryEditor) {
      // We don't use editorContext.setOperationName because that function does not accept null values for some reason
      // Note to future me - if you need to look at the source of setOperationName, search everywhere for `'setOperationName'` in the graphiql codebase
      // NOTE: I'm not sure if setting this value actually does anything
      queryEditor.operationName = newOperationName ?? null;
    }
    // by updating the active tab values, we are able to switch the "active operation" to whatever the user has just typed out
    editorContext?.updateActiveTabValues({operationName: newOperationName})
    onChange({ ...query, operationName: newOperationName });
  };

  const onParsingOptionChange = (event: React.FormEvent<HTMLInputElement>, field: keyof ParsingOption, editedIndex: number) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => index === editedIndex
        ? {
          ...parsingOption,
          [field]: event.currentTarget.value,
        }
        : parsingOption
      )
    });
  };

  const deleteParsingOption = (index: number) => {
    const newParsingOptions: ParsingOption[] = [];
    newParsingOptions.push(...query.parsingOptions.slice(0, index));
    newParsingOptions.push(...query.parsingOptions.slice(index + 1, query.parsingOptions.length));
    onChange({
      ...query,
      parsingOptions: newParsingOptions,
    })
  };
  const newParsingOption = () => {
    const newParsingOptions = [...query.parsingOptions];
    const timePath = query.parsingOptions.length === 0 ? "time.path" : query.parsingOptions[query.parsingOptions.length - 1].timePath;
    newParsingOptions.push({
      "dataPath": "data.path",
      "timePath": timePath,
    })
    onChange({
      ...query,
      parsingOptions: newParsingOptions,
    })
  };

  const currentOperationName = editorContext?.queryEditor?.operationName;
  useEffect(() => {
    // if currentOperationName is null, that means that the query is unnamed
    // currentOperationName should never be undefined unless queryEditor is undefined
    if (currentOperationName !== undefined && query.operationName !== currentOperationName) {
      // Remember that in our world, we use the string | undefined type for operationName,
      //   so we're basically converting null to undefined here
      onChange({ ...query, operationName: currentOperationName ?? undefined });
    }
  }, [onChange, query, currentOperationName]);


  return (
    <>
      <h3 className="page-heading">Query</h3>
      <div className="gf-form-group">
        <div className="gf-form" style={{height: "450px"}}>
          {/*TODO allow this to be resized*/}
          <GraphiQLInterface
            showPersistHeadersSettings={false}
            disableTabs={true}
            isHeadersEditorEnabled={false} // TODO consider enabling customizable headers later
            onEditQuery={(value) => {
              onChange({...query, queryText: value});
            }}
            onEditVariables={(variablesJsonString) => {
              if (variablesJsonString) {
                onChange({...query, variables: variablesJsonString});
              } else {
                onChange({...query, variables: undefined});
              }
            }}
          />
        </div>
        <div className="gf-form-inline">
          <InlineField label="Operation Name" labelWidth={LABEL_WIDTH}
                       tooltip="The operationName passed to the GraphQL endpoint. This can be left blank unless you specify multiple queries.">
            <Input
              onChange={onOperationNameChange} value={query.operationName ?? ''}
              width={INPUT_WIDTH}
            />
          </InlineField>
        </div>
        {query.parsingOptions.map((parsingOption, index) => <>
          <div className="gf-form-inline" style={{marginTop: "1em"}}>
            <InlineField label={`Parsing Option ${index + 1}`} labelWidth={LABEL_WIDTH}>
              <div></div>
            </InlineField>
            {query.parsingOptions.length === 1
              ? null
              : <IconButton name={"trash-alt"} onClick={() => deleteParsingOption(index)}/>
            }
          </div>
          <div className="gf-form-inline">
            <InlineField label="Data Path" labelWidth={LABEL_WIDTH}
                         tooltip="Dot-delimited path to an array nested in the root of the JSON response.">
              <Input
                onChange={event => onParsingOptionChange(event, "dataPath", index)}
                value={parsingOption.dataPath ?? ''}
                width={INPUT_WIDTH}/>
            </InlineField>
          </div>
          <div className="gf-form-inline">
            <InlineField label="Time Path" labelWidth={LABEL_WIDTH}
                         tooltip="Dot-delimited path to the time field relative to the data path">
              <Input
                onChange={event => onParsingOptionChange(event, "timePath", index)}
                value={parsingOption.timePath ?? ''}
                width={INPUT_WIDTH}/>
            </InlineField>
          </div>
        </>)}

        {/*https://developers.grafana.com/ui/latest/index.html?path=/docs/buttons-button--examples*/}
        {/*https://grafana.com/developers/saga/Components/Buttons/Button*/}
        <Button
          variant="secondary"
          style={{marginTop: "1em"}}
          onClick={() => newParsingOption()}
        >
          Add Parsing Option
        </Button>
      </div>
    </>
  );

}
