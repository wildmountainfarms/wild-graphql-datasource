import React, { ChangeEvent, KeyboardEvent, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { Button, Checkbox, Combobox, IconButton, InlineField, Input, TextArea, useTheme2 } from '@grafana/ui';
import { CoreApp, QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import {
  DEFAULT_LABEL_OPTION_FIELD_CONFIG,
  getQueryVariablesAsJsonString,
  LabelOption,
  LabelOptionType,
  ParsingOption,
  TimeField,
  WildGraphQLAnyQuery,
  WildGraphQLDataSourceOptions,
} from '../types';
import { GraphiQLInterface } from 'graphiql';
import {
  GraphiQLProvider,
  useGraphiQL,
  useGraphiQLActions,
} from '@graphiql/react';
import { DOC_EXPLORER_PLUGIN, DocExplorerStore } from '@graphiql/plugin-doc-explorer';
import { explorerPlugin } from '@graphiql/plugin-explorer';
import type { Fetcher, FetcherOpts, FetcherParams, Storage } from '@graphiql/toolkit';
import { getBackendSrv, getTemplateSrv } from '@grafana/runtime';
import { firstValueFrom } from 'rxjs';

import '../monacoWorkers';
import 'graphiql/style.css';
import '@graphiql/plugin-explorer/style.css';
import './modify_graphiql.css';
import { ExecutionResult } from 'graphql';

import { getInterpolatedAutoPopulatedVariables, interpolateVariables } from '../variables';

export type Props = QueryEditorProps<DataSource, WildGraphQLAnyQuery, WildGraphQLDataSourceOptions>;
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
function createFetcher(url: string, withCredentials: boolean, basicAuth?: string): Fetcher {
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
      ...interpolateVariables(graphQLParams.variables ?? {}, templateSrv), // remember one of the downsides here is that we cannot pass scopedVars here because we don't have access to it
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

  const noopStorage = useMemo<Storage>(() => ({
    getItem: () => null,
    setItem: () => {},
    removeItem: () => {},
    clear: () => {},
    length: 0,
  }), []);
  const plugins = useMemo(() => [DOC_EXPLORER_PLUGIN, explorerPlugin()], []);

  return (
    <GraphiQLProvider
      storage={noopStorage}
      fetcher={fetcher}
      initialQuery={correctedQuery.queryText}
      initialVariables={getQueryVariablesAsJsonString(correctedQuery)}
      plugins={plugins}
      referencePlugin={DOC_EXPLORER_PLUGIN}
    >
      <DocExplorerStore>
        {/* Hide execution during alerting because the to and from variables are not populated correctly. */}
        <div className={isAlerting ? "hide-execute-button" : ""}>
          <InnerQueryEditor
            query={correctedQuery}
            onChange={props.onChange}
            app={props.app}
          />
        </div>
      </DocExplorerStore>
    </GraphiQLProvider>
  );
}

export default QueryEditor;

function InnerQueryEditor({ query, onChange, app }: InnerQueryProps) {
  const isBackendOnlyQuery = app === CoreApp.CloudAlerting || app === CoreApp.UnifiedAlerting;
  const currentOperationName = useGraphiQL((state) => state.operationName);
  const hasParsedOperations = useGraphiQL((state) => state.operations !== undefined);
  const { setOperationName } = useGraphiQLActions();
  const labelToAddRef = useRef<HTMLInputElement>(null);
  const theme = useTheme2();
  const editorDialogRef = useRef<HTMLDivElement>(null);
  const [isEditorExpanded, setEditorExpanded] = useState(false);

  useEffect(() => {
    const containEscape = (event: globalThis.KeyboardEvent) => {
      if (event.key !== 'Escape' || !(event.target instanceof Element)) {
        return;
      }
      if (event.target.closest('.graphiql-dialog, .graphiql-dropdown-content')) {
        // Radix also listens on document in the capture phase. Let its listener
        // dismiss the popup, but stop Escape before Grafana's bubbling shortcut.
        // React handlers run too late: Radix can already have unmounted the popup.
        event.stopPropagation();
        return;
      }
      if (isEditorExpanded && !editorDialogRef.current?.contains(event.target)) {
        // Closing a GraphiQL popup can leave focus on document.body without a
        // focusin event. Escape must still close the expanded editor first.
        event.stopPropagation();
        event.preventDefault();
        setEditorExpanded(false);
      }
    };
    document.addEventListener('keydown', containEscape, true);
    return () => document.removeEventListener('keydown', containEscape, true);
  }, [isEditorExpanded]);

  useEffect(() => {
    if (!isEditorExpanded) {
      return;
    }

    const dialog = editorDialogRef.current!;
    const previousFocus = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    let tabBackwards = false;
    const focusableElements = () => Array.from(dialog.querySelectorAll<HTMLElement>(
      'button, input, textarea, select, a[href], [tabindex]'
    )).filter((element) => element.tabIndex >= 0 && !element.matches(':disabled') && element.getClientRects().length > 0);
    const rememberTabDirection = (event: globalThis.KeyboardEvent) => {
      if (event.key === 'Tab') {
        tabBackwards = event.shiftKey;
      }
    };
    const containFocus = (event: FocusEvent) => {
      const target = event.target;
      if (!(target instanceof Element) || dialog.contains(target)) {
        return;
      }
      // GraphiQL renders these controls in portals outside the editor's DOM tree.
      if (target.closest('.graphiql-dialog, .graphiql-dropdown-content')) {
        return;
      }
      const elements = focusableElements();
      (tabBackwards ? elements[elements.length - 1] : elements[0])?.focus();
    };
    document.body.style.overflow = 'hidden';
    document.body.classList.add('wild-graphql-editor-expanded');
    document.addEventListener('keydown', rememberTabDirection, true);
    document.addEventListener('focusin', containFocus);
    focusableElements()[0]?.focus();
    return () => {
      document.removeEventListener('keydown', rememberTabDirection, true);
      document.removeEventListener('focusin', containFocus);
      document.body.classList.remove('wild-graphql-editor-expanded');
      document.body.style.overflow = previousOverflow;
      previousFocus?.focus({ preventScroll: true });
    };
  }, [isEditorExpanded]);

  // GraphiQL 5 registers its query-change listener when Monaco initializes and
  //   retains that render's onEditQuery callback.
  //   Spreading the captured query can overwrite newer fields, including operationName.
  //   Read the latest committed props through a ref so even the retained callback uses current values.
  //   Upstream should refresh the onEdit listener when the callback changes,
  //   separately from editor creation, or forward calls to the latest callback.
  const latestQueryPropsRef = useRef({ query, onChange });
  useLayoutEffect(() => {
    latestQueryPropsRef.current = { query, onChange };
  }, [query, onChange]);
  const onEditQuery = useCallback((value: string) => {
    const latest = latestQueryPropsRef.current;
    latest.onChange({ ...latest.query, queryText: value });
  }, []);

  const onOperationNameChange = (event: ChangeEvent<HTMLInputElement>) => {
    const newOperationName = event.target.value || undefined;
    setOperationName(newOperationName ?? "");
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

  useEffect(() => {
    // if currentOperationName is null, that means that the query is unnamed
    // Treat an empty, null, or undefined operation name the same.
    //   We need to do this because otherwise we are constantly doing onChange calls, which results in 100% CPU utilization
    if (
      hasParsedOperations
      && (query.operationName || undefined) !== (currentOperationName || undefined)
    ) {
      // Remember that in our world, we use the string | undefined type for operationName,
      //   so we're basically converting null to undefined here
      onChange({ ...query, operationName: currentOperationName || undefined });
    }
  }, [onChange, query, currentOperationName, hasParsedOperations]);

  return (
    <>
      <h3 className="page-heading">Query</h3>
      <div className="gf-form-group">
        {isEditorExpanded && <div
          className="wild-graphql-editor-backdrop"
          aria-hidden="true"
          onClick={() => setEditorExpanded(false)}
        />}
        <div
          ref={editorDialogRef}
          className={`wild-graphql-editor-dialog${isEditorExpanded ? " is-expanded" : ""}`}
          style={{ background: theme.colors.background.primary }}
          role={isEditorExpanded ? 'dialog' : 'presentation'}
          aria-label={isEditorExpanded ? 'Query editor' : undefined}
          aria-modal={isEditorExpanded || undefined}
          onKeyDown={(event) => {
            if (isEditorExpanded && event.key === 'Escape' && event.currentTarget.contains(event.target as Node)) {
              // Monaco gets first chance to dismiss suggestions. Keep Escape
              // from also triggering Grafana's shortcut to leave the panel editor.
              event.stopPropagation();
              if (!event.defaultPrevented) {
                event.preventDefault();
                setEditorExpanded(false);
              }
            }
          }}
        >
          <div className="wild-graphql-editor-toolbar">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setEditorExpanded(!isEditorExpanded)}
            >
              {isEditorExpanded ? 'Close expanded editor' : 'Expand editor'}
            </Button>
          </div>
          <div className="gf-form wild-graphql-editor-container">
            <GraphiQLInterface
              className="wild-graphql-query-editor"
              showPersistHeadersSettings={false}
              isHeadersEditorEnabled={false} // TODO consider enabling customizable headers later
              onEditQuery={onEditQuery}
              onEditVariables={(variablesJsonString) => {
                if (variablesJsonString.trimStart()) {
                  onChange({...query, variables: variablesJsonString});
                } else {
                  onChange({...query, variables: undefined});
                }
              }}
            />
          </div>
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
                    <Combobox
                      width={14}
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
                    <>
                      <InlineField label="If absent" labelWidth={10}>
                        <Combobox
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
                                  ...(fieldConfig!),
                                  required: newValue === "required",
                                  defaultValue: newValue === "omit" ? undefined : (fieldConfig!.defaultValue ?? "")
                                }
                              });
                            }
                          }}
                        />
                      </InlineField>
                      {fieldConfigSelection === "default" &&
                        <InlineField label="Default" labelWidth={10}>
                          <Input
                            width={INPUT_WIDTH}
                            value={fieldConfig!.defaultValue!}
                            onChange={(event) => {
                              setLabelOption(parsingOptionIndex, labelOptionIndex, {
                                ...labelOption,
                                fieldConfig: {
                                  ...(fieldConfig!),
                                  required: false,
                                  defaultValue: event.currentTarget.value
                                }
                              });
                            }}
                          />
                        </InlineField>
                      }
                      <InlineField
                        label="Frame exclude"
                        labelWidth={16}
                        tooltip="When checked, this field will not be included in the data frame."
                      >
                        <Checkbox
                          label=""
                          value={fieldConfig!.excludeFieldFromDataFrame === true}
                          onChange={(event) => {
                            setLabelOption(parsingOptionIndex, labelOptionIndex, {
                              ...labelOption,
                              fieldConfig: {
                                ...(fieldConfig!),
                                excludeFieldFromDataFrame: event.currentTarget.checked || undefined // prefer using undefined rather than false value
                              }
                            })
                          }}
                        />
                      </InlineField>
                    </>
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
