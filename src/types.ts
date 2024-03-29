import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

type VariablesType = string | Record<string, any>;

export interface TimeField {
  timePath: string;
  // TODO add a time format option here
}

export interface LabelOption {
  name: string;
  type: LabelOptionType;
  /** When {@link type} is {@link LabelOptionType.CONSTANT}, this represents a text value that is constant.
   * When {@link type} is {@link LabelOptionType.FIELD}, this represents the path to a field relative to the data path */
  value: string;
}
export enum LabelOptionType {
  CONSTANT = "constant",
  FIELD = "field",
}

export interface ParsingOption {
  // TODO make sure an empty string is valid in the backend like this says
  /** Required. The path to the array of data or object of data. Note: An empty string is valid -- it refers to the top-most data */
  dataPath: string;
  /**
   * The path to the time. An undefined value or an empty array means that no fields will be interpreted as time fields.
   *
   * If one or multiple of the time paths are blank, those blank time paths should not be treated as time paths and may be disregarded during processing.
   */
  timeFields?: TimeField[];
  /** The label options. The number of label options and the names of the label options should be consistent between parsing options for the best user experience.*/
  labelOptions?: LabelOption[];
}


interface WildGraphQLCommonQuery extends DataQuery {
  queryText: string;
  /** The operation name if explicitly set. Note that an empty string should be treated the same way as an undefined value, although storing an undefined value is preferred.*/
  operationName?: string;
  variables?: VariablesType;

  parsingOptions: ParsingOption[];
}

export function getQueryVariablesAsJsonString(query: WildGraphQLCommonQuery): string {
  const variables = query.variables;
  if (variables === undefined) {
    return "";
  }
  if (typeof variables === 'string') {
    return variables;
  }
  return JSON.stringify(variables, null, 2); // TODO consider allowing customization of size of tabs used
}
export function getQueryVariablesAsJson(query: WildGraphQLCommonQuery): Record<string, any> {
  const variables = query.variables;
  if (variables === undefined) {
    return {};
  }
  if (typeof variables === 'string') {
    try {
      return JSON.parse(variables);
    } catch (e) {
      if (e instanceof SyntaxError) {
        console.error("Got a syntax error while parsing variables!", e);
        console.log("Variables is", variables);
        return {}; // in the case of a syntax error, let's just make it so variables are not included
      } else {
        throw e; // unexpected exception
      }
    }
  }
  return variables;
}

/**
 * This interface represents the main type of query, which is used for most queries in Grafana
 */
export interface WildGraphQLMainQuery extends WildGraphQLCommonQuery {
}

export interface WildGraphQLAnnotationQuery extends WildGraphQLCommonQuery {
}

/** This type represents the possible options that can be stored in the datasource JSON for queries */
export type WildGraphQLAnyQuery = (WildGraphQLMainQuery | WildGraphQLAnnotationQuery) &
  Partial<WildGraphQLMainQuery> &
  Partial<WildGraphQLAnnotationQuery>;


/**
 * These are options configured for each DataSource instance
 */
export interface WildGraphQLDataSourceOptions extends DataSourceJsonData {
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface WildGraphQLSecureJsonData {
  // TODO We should support secret fields that can be passed to GraphQL queries as arguments
}

// TODO don't use Long here
export const DEFAULT_QUERY: Partial<WildGraphQLMainQuery> = {
  queryText: `query ($from: Long!, $to: Long!) {
  queryStatus(from: $from, to: $to) {
    batteryVoltage {
      dateMillis
      packet {
        batteryVoltage
      }
    }
  }
}
`,
  parsingOptions: [
    {
      dataPath: "queryStatus.batteryVoltage",
      timeFields: [{ timePath: "dateMillis" }]
    }
  ]
};

export const DEFAULT_ALERTING_QUERY: Partial<WildGraphQLMainQuery> = {
  queryText: `query ($from: Long!, $to: Long!) {
  queryStatus(sourceId: "default", from: $from, to: $to) {
    batteryVoltage {
      dateMillis
      packet {
        batteryVoltage
      }
    }
  }
}
`,
  parsingOptions: [
    {
      dataPath: "queryStatus.batteryVoltage",
      timeFields: [{ timePath: "dateMillis" }]
    }
  ]
};

export const DEFAULT_ANNOTATION_QUERY: Partial<WildGraphQLAnnotationQuery> = {
  queryText: `query ($from: Long!, $to: Long!) {
  queryEvent(from:$from, to:$to) {
    mateCommand {
      dateMillis
      packet {
        commandName
      }
    }
  }
}
`,
  parsingOptions: [
    {
      dataPath: "queryEvent.mateCommand",
      timeFields: [{ timePath: "dateMillis" }]
    }
  ]
};
