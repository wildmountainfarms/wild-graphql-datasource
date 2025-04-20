import React, {ChangeEvent, KeyboardEvent, useEffect, useMemo, useRef} from 'react';
import {Button, Checkbox, IconButton, InlineField, Input, Select, TextArea} from '@grafana/ui';
import {CoreApp, QueryEditorProps} from '@grafana/data';
import {DataSource} from '../datasource';
import {
  DEFAULT_LABEL_OPTION_FIELD_CONFIG,
  getQueryVariablesAsJsonString,
  LabelOption,
  LabelOptionType,
  ParsingOption,
  TimeField,
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
interface InnerQueryProps {
  query: WildGraphQLAnyQuery
  onChange: (value: WildGraphQLAnyQuery) => void,
  app?: CoreApp
}

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

  // *sometimes* and only sometimes when creating a new panel the query won't be populated with the default query.
  //   When that happens any assumption we make about the presence of fields of query, we get an NPE.
  //   So although these default values aren't ideal,
  //   we use them here because we don't need to replicate default query logic here, as if this happens it's for Grafana to fix
  const correctedQuery: WildGraphQLAnyQuery = {
    refId: "", // I don't think there's a documented case of refId not being present, but we'll guard against it anyway
    queryText: "",
    parsingOptions: [],
    ...(query as Partial<WildGraphQLAnyQuery>), // cast to partial to make compiler point out missing fields
  };

  return (
    <>
      {/*By not providing storage, history contexts, they won't be used*/}
      {/*<StorageContextProvider storage={DummyStorage}>*/}
      {/*  <HistoryContextProvider maxHistoryLength={0}>*/}
      <EditorContextProvider
        // defaultQuery is the query that is used for new tabs, but we already define the open tabs here
        defaultTabs={[{
          query: correctedQuery.queryText,
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
                    query={correctedQuery}
                    onChange={props.onChange}
                    app={props.app}
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

function InnerQueryEditor({ query, onChange, app }: InnerQueryProps) {
  const isBackendOnlyQuery = app === CoreApp.CloudAlerting || app === CoreApp.UnifiedAlerting;
  const editorContext = useEditorContext();
  const labelToAddRef = useRef<HTMLInputElement>(null);

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

  const setParsingOption = (parsingOptionIndex: number, newParsingOption: ParsingOption) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => index === parsingOptionIndex
        ? newParsingOption
        : parsingOption
      )
    });
  };
  const updateParsingOptionArray = <
    K extends keyof ParsingOption,
    T extends ParsingOption[K],
  >(
    arrayKey: K,
    parsingOptionIndex: number,
    itemIndex: number,
    // TODO figure out how to do generics better here
    // @ts-ignore
    newItem: ParsingOption[K][number],
) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => {
        if (index !== parsingOptionIndex) {
          return parsingOption;
        }

        // TODO note that T[] is not technically correct right here
        const currentArray = parsingOption[arrayKey] as T[] | undefined;

        if (!currentArray) {
          return {
            ...parsingOption,
            [arrayKey]: [newItem]
          };
        }

        //
        if (itemIndex >= currentArray.length) {
          return {
            ...parsingOption,
            [arrayKey]: [...currentArray, newItem]
          };
        }

        return {
          ...parsingOption,
          [arrayKey]: currentArray.map((item, i) => i === itemIndex ? newItem : item)
        };
      })
    });
  };
  const setExplodeArrayPath = (parsingOptionIndex: number, explodeArrayPathsIndex: number, newExplodeArrayPath: string) => {
    updateParsingOptionArray("explodeArrayPaths", parsingOptionIndex, explodeArrayPathsIndex, newExplodeArrayPath);
  };
  const setTimeField = (parsingOptionIndex: number, timeFieldIndex: number, newTimeField: TimeField) => {
    updateParsingOptionArray("timeFields", parsingOptionIndex, timeFieldIndex, newTimeField);
  };
  const setLabelOption = (parsingOptionIndex: number, labelOptionIndex: number, newLabelOption: LabelOption) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => index === parsingOptionIndex
        ? {
          ...parsingOption,
          labelOptions: parsingOption.labelOptions!.map((labelOption, index) => index === labelOptionIndex
            ? newLabelOption
            : labelOption
          )
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
    });
  };
  const swapParsingOption = (index1: number, index2: number) => {
    const newParsingOptions: ParsingOption[] = [...query.parsingOptions];
    const temp = newParsingOptions[index1];
    newParsingOptions[index1] = newParsingOptions[index2];
    newParsingOptions[index2] = temp;
    onChange({
      ...query,
      parsingOptions: newParsingOptions
    })
  };
  const deleteLabelOption = (parsingOptionIndex: number, labelOptionIndex: number) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => {
          if (index === parsingOptionIndex) {
            const newLabelOptions: LabelOption[] = [];
            newLabelOptions.push(...parsingOption.labelOptions!.slice(0, labelOptionIndex));
            newLabelOptions.push(...parsingOption.labelOptions!.slice(labelOptionIndex + 1, parsingOption.labelOptions!.length));
            return {
              ...parsingOption,
              labelOptions: newLabelOptions || undefined,
            };
          }
          return parsingOption;
        }
      )
    });
  };
  const deleteExplodeArrayPath = (parsingOptionIndex: number, explodeArrayPathIndex: number) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => {
          if (index === parsingOptionIndex) {
            const newExplodeArrayPaths: string[] = [];
            newExplodeArrayPaths.push(...parsingOption.explodeArrayPaths!.slice(0, explodeArrayPathIndex));
            newExplodeArrayPaths.push(...parsingOption.explodeArrayPaths!.slice(explodeArrayPathIndex + 1, parsingOption.explodeArrayPaths!.length));
            return {
              ...parsingOption,
              explodeArrayPaths: newExplodeArrayPaths,
            };
          }
          return parsingOption;
        }
      )
    });
  }
  const deleteTimeField = (parsingOptionIndex: number, timeFieldIndex: number) => {
    onChange({
      ...query,
      parsingOptions: query.parsingOptions.map((parsingOption, index) => {
          if (index === parsingOptionIndex) {
            const newTimeFields: TimeField[] = [];
            newTimeFields.push(...parsingOption.timeFields!.slice(0, timeFieldIndex));
            newTimeFields.push(...parsingOption.timeFields!.slice(timeFieldIndex + 1, parsingOption.timeFields!.length));
            return {
              ...parsingOption,
              timeFields: newTimeFields,
            };
          }
          return parsingOption;
        }
      )
    });
  };

  const addNewParsingOption = () => {
    const newParsingOptions = [...query.parsingOptions];
    const lastParsingOption = query.parsingOptions.length === 0
      ? undefined
      : query.parsingOptions[query.parsingOptions.length - 1];
    const timePaths = lastParsingOption === undefined ? undefined : lastParsingOption.timeFields;
    const labelOptions = lastParsingOption === undefined
      ? undefined
      : lastParsingOption?.labelOptions?.map(labelOption => ({
        name: labelOption.name,
        type: LabelOptionType.CONSTANT,
        value: ""
      }));
    newParsingOptions.push({
      "dataPath": "data.path",
      "timeFields": timePaths,
      labelOptions: labelOptions || undefined
    });
    onChange({
      ...query,
      parsingOptions: newParsingOptions,
    });
  };

  const addNewLabel = () => {
    const value = labelToAddRef.current?.value;
    if (value === undefined) {
      console.error("Label to add has an uninitialized ref!")
    } else {
      labelToAddRef.current!.value = "";
      const newParsingOptions = query.parsingOptions.map((parsingOption) => {
        if (parsingOption.labelOptions?.find((labelOption) => labelOption.name === value) !== undefined) {
          // if this parsing option already has a label option with the same name, don't add it
          return parsingOption;
        }
        const newLabelOptions = [...(parsingOption.labelOptions ?? [])];
        newLabelOptions.push({
          name: value,
          type: LabelOptionType.CONSTANT,
          value: "",
        });
        return {
          ...parsingOption,
          labelOptions: newLabelOptions,
        };
      });
      onChange({
        ...query,
        parsingOptions: newParsingOptions,
      });
    }
  };

  const handleLabelToAddKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Enter') {
      addNewLabel();
    }
  };
  // const cleanUpTimePaths = () => {
  //   onChange({
  //     ...query,
  //     parsingOptions: query.parsingOptions.map(parsingOption => ({
  //       ...parsingOption,
  //       timePaths: parsingOption.timePaths?.filter(timePath => timePath != "") || undefined
  //     }))
  //   });
  // };

  const currentOperationName = editorContext?.queryEditor?.operationName;
  useEffect(() => {
    // if currentOperationName is null, that means that the query is unnamed
    // currentOperationName should never be undefined unless queryEditor is undefined
    // Treat an empty, null, or undefined operation name the same.
    //   We need to do this because otherwise we are constantly doing onChange calls, which results in 100% CPU utilization
    if (
      currentOperationName !== undefined
      && (query.operationName || undefined) !== (currentOperationName || undefined)
    ) {
      // Remember that in our world, we use the string | undefined type for operationName,
      //   so we're basically converting null to undefined here
      onChange({ ...query, operationName: currentOperationName || undefined });
    }
  }, [onChange, query, currentOperationName]);


  return (
    <>
      <h3 className="page-heading">Query</h3>
      <div className="gf-form-group">
        {/*When we attempt to debug the Combobox error, this line reproduces it*/}
        {/*<Combobox options={[{label: "lol1", value: "lol1"}, {label: "lol2", value: "lol2"}]} onChange={newValue => {}} width={"auto"} minWidth={10}/>*/}
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
              if (variablesJsonString.trimStart()) {
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
        {!isBackendOnlyQuery && <>
          <Checkbox
            label="Define Advanced Variables JSON"
            value={query.variablesWithFullInterpolation !== undefined}
            onChange={(event) => {
              onChange({
                ...query,
                variablesWithFullInterpolation: event.currentTarget.checked ? "{\n  \n}" : undefined
              })
            }}
          />
          {query.variablesWithFullInterpolation !== undefined &&
            <TextArea
              style={{
                minHeight:"10em"
              }}
              value={query.variablesWithFullInterpolation}
              onChange={(event) => {
                onChange({
                  ...query,
                  variablesWithFullInterpolation: event.currentTarget.value
                })
              }}
            />
          }
        </>}
      </div>
      <h3 className="page-heading">Parsing Options</h3>
      <div className="gf-form-group">
        {query.parsingOptions.map((parsingOption, parsingOptionIndex) => {
          const displayedExplodeArrayPaths = [
            ...(parsingOption.explodeArrayPaths ?? []),
            ""
          ];
          const displayedTimeFields = [
            ...(parsingOption.timeFields ?? []),
            {
              timePath: ""
            }
          ];
          return <>
            <div className="gf-form-inline" style={{marginTop: "1em"}}>
              <InlineField label={`Parsing Option ${parsingOptionIndex + 1}`} labelWidth={LABEL_WIDTH}>
                <div></div>
              </InlineField>
              {parsingOptionIndex !== 0 &&
                <IconButton
                  name={"arrow-up"}
                  aria-label="Move up"
                  onClick={() => swapParsingOption(parsingOptionIndex, parsingOptionIndex - 1)}
                />
              }
              {parsingOptionIndex < query.parsingOptions.length - 1 &&
                <IconButton
                  name={"arrow-down"}
                  aria-label="Move down"
                  onClick={() => swapParsingOption(parsingOptionIndex, parsingOptionIndex + 1)}
                />
              }
              {query.parsingOptions.length !== 1 &&
                <IconButton
                  name={"trash-alt"}
                  aria-label="Remove"
                  onClick={() => deleteParsingOption(parsingOptionIndex)}
                />
              }
            </div>
            <div className="gf-form-inline">
              <InlineField label="Data Path" labelWidth={LABEL_WIDTH}
                           tooltip="Dot-delimited path to an array nested in the root of the JSON response.">
                <Input
                  onChange={event => setParsingOption(parsingOptionIndex, {
                    ...parsingOption,
                    dataPath: event.currentTarget.value
                  })}
                  value={parsingOption.dataPath ?? ''}
                  width={INPUT_WIDTH}/>
              </InlineField>
            </div>
            {displayedExplodeArrayPaths.map((explodeArrayPath, explodeArrayPathIndex) => <>
              <div className="gf-form-inline">
                <InlineField
                  label={explodeArrayPathIndex === displayedExplodeArrayPaths.length - 1 ? "Add Explode Array Path" : "Explode Array Path"}
                  labelWidth={LABEL_WIDTH}
                  tooltip="Dot-delimited path to arrays within the response to explode to make multiple rows, rather than multiple columns."
                >
                  <Input
                    onChange={event => setExplodeArrayPath(parsingOptionIndex, explodeArrayPathIndex, event.currentTarget.value)}
                    value={explodeArrayPath}
                    onBlur={event => {
                      if (explodeArrayPathIndex !== displayedExplodeArrayPaths.length - 1 && explodeArrayPath === "") {
                        deleteExplodeArrayPath(parsingOptionIndex, explodeArrayPathIndex);
                      }
                    }}
                    width={INPUT_WIDTH}/>
                </InlineField>
                {explodeArrayPathIndex !== displayedExplodeArrayPaths.length - 1 &&
                  <IconButton
                    name={"minus"}
                    aria-label="Remove explode array path"
                    onClick={() => deleteExplodeArrayPath(parsingOptionIndex, explodeArrayPathIndex)}
                  />
                }
              </div>
            </>)}
            {displayedTimeFields.map((timeField, timeFieldIndex) => <>
              <div className="gf-form-inline">
                <InlineField
                  label={timeFieldIndex === displayedTimeFields.length - 1 ? "Add Time Path" : "Time Path"}
                  labelWidth={LABEL_WIDTH}
                  tooltip="Dot-delimited path to the time field relative to the data path"
                >
                  <Input
                    onChange={event => setTimeField(parsingOptionIndex, timeFieldIndex, {
                      ...timeField,
                      timePath: event.currentTarget.value
                    })}
                    value={timeField.timePath}
                    onBlur={event => {
                      if (timeFieldIndex !== displayedTimeFields.length && timeField.timePath === "") {
                        deleteTimeField(parsingOptionIndex, timeFieldIndex);
                      }
                    }}
                    width={INPUT_WIDTH}/>
                </InlineField>
                {/*TODO add time format option here*/}
                {timeFieldIndex !== displayedTimeFields.length - 1 &&
                  <IconButton
                    name={"minus"}
                    aria-label="Remove time path"
                    onClick={() => deleteTimeField(parsingOptionIndex, timeFieldIndex)}
                  />
                }
              </div>
            </>)}
            {parsingOption.labelOptions?.map((labelOption, labelOptionIndex) => {
              // fieldConfig and fieldConfigSelection are undefined ONLY when labelOption.type is CONSTANT
              const fieldConfig = labelOption.type === LabelOptionType.CONSTANT
                ? undefined
                : (labelOption.fieldConfig ?? DEFAULT_LABEL_OPTION_FIELD_CONFIG);
              const fieldConfigSelection = fieldConfig === undefined
                ? undefined
                : fieldConfig.required
                  ? "required"
                  : fieldConfig.defaultValue === undefined ? "omit" : "default";
              return <>
                <div className="gf-form-inline">
                  <InlineField
                    label={`Label: "${labelOption.name}"`}
                    tooltip={`Specify how the custom label "${labelOption.name}" should be populated. A type of "Constant" means that you may put whatever text you would like as the label. A type of "Field" means that the given field will be used as the label's value.`}
                    labelWidth={LABEL_WIDTH}
                  >
                    {/* TODO use Combobox here https://developers.grafana.com/ui/latest/index.html?path=/docs/forms-combobox--docs*/}
                    <Select
                      width={16}
                      options={[
                        {label: "Constant", value: LabelOptionType.CONSTANT},
                        {label: "Field", value: LabelOptionType.FIELD},
                      ]}
                      value={labelOption.type}
                      onChange={(value) => {
                        const newType = value.value;
                        if (newType !== undefined) {
                          setLabelOption(parsingOptionIndex, labelOptionIndex, {
                            ...labelOption,
                            type: newType,
                          });
                        }
                      }}
                    />

                  </InlineField>
                  <InlineField label="Value" labelWidth={8}>
                    <Input
                      width={INPUT_WIDTH}
                      value={labelOption.value}
                      onChange={(event) => {
                        setLabelOption(parsingOptionIndex, labelOptionIndex, {
                          ...labelOption,
                          value: event.currentTarget.value,
                        })
                      }}
                    />
                  </InlineField>

                  {fieldConfig &&
                    <InlineField label="If absent" labelWidth={10}>
                      <Select
                        width={16}
                        options={[
                          {label: "Error", value: "required"},
                          {label: "Omit", value: "omit"},
                          {label: "Use default", value: "default"},
                        ]}
                        value={fieldConfigSelection!}
                        onChange={(value) => {
                          const newValue = value.value;
                          if (newValue !== undefined) {
                            setLabelOption(parsingOptionIndex, labelOptionIndex, {
                              ...labelOption,
                              fieldConfig: {
                                required: newValue === "required",
                                defaultValue: newValue === "omit" ? undefined : (fieldConfig!.defaultValue ?? "")
                              }
                            });
                          }
                        }}
                      />
                    </InlineField>
                  }
                  {fieldConfigSelection === "default" &&
                    <InlineField label="Default" labelWidth={10}>
                      <Input
                        width={INPUT_WIDTH}
                        value={fieldConfig!.defaultValue!}
                        onChange={(event) => {
                          setLabelOption(parsingOptionIndex, labelOptionIndex, {
                            ...labelOption,
                            fieldConfig: {
                              required: false,
                              defaultValue: event.currentTarget.value
                            }
                          })
                        }}
                      />
                    </InlineField>
                  }
                  <IconButton
                    name={"minus"}
                    aria-label="Remove"
                    onClick={() => deleteLabelOption(parsingOptionIndex, labelOptionIndex)}
                  />
                </div>
              </>;
            })}
          </>;
        })}

        {/*https://developers.grafana.com/ui/latest/index.html?path=/docs/buttons-button--examples*/}
        {/*https://grafana.com/developers/saga/Components/Buttons/Button*/}
        <Button
          variant="secondary"
          style={{marginTop: "1em"}}
          onClick={() => addNewParsingOption()}
        >
          Add Parsing Option
        </Button>
        <div className="gf-form-inline" style={{marginTop: "0.5em"}}>
          <InlineField label="Create label" labelWidth={LABEL_WIDTH}
                       tooltip="Type the name of the label you would like to add, then press the plus button.">
            <Input
              ref={labelToAddRef}
              onKeyDown={handleLabelToAddKeyDown}
              defaultValue=''
              width={INPUT_WIDTH}/>
          </InlineField>
          <IconButton
            name={"plus"}
            aria-label="New label"
            onClick={() => addNewLabel()}
          />
        </div>
      </div>
    </>
  );

}
