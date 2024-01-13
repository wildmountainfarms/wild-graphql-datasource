import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

type VariablesType = string | Record<string, any>;


export interface ParsingOption {
  dataPath: string;
  /** Required. The path to the time (this represents the "start time" in the case when {@link timeEndPath} is defined) */
  timePath: string;
  // TODO use timeEndPath on the backend
  /** Optional. The path to the "end time". Should only be shown for the annotation query. A blank string should be treated the same as undefined*/
  timeEndPath?: string;
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
        return {}; // in the case of a syntax error, let's just make it so variables are not included
        // TODO consider logging an error here or something
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

export const DEFAULT_QUERY: Partial<WildGraphQLMainQuery> = {
  queryText: `query BatteryVoltage($sourceId: String!, $from: Long!, $to: Long!) {
  queryStatus(sourceId: $sourceId, from: $from, to: $to) {
    batteryVoltage {
      dateMillis
      fragmentId
      packet {
        batteryVoltage
        identifier {
          representation
        }
        identityInfo {
          displayName
        }
      }
    }
  }
}
`,
  variables: {
    "sourceId": "default"
  },
  parsingOptions: [
    {
      dataPath: "queryStatus.batteryVoltage",
      timePath: "dateMillis"
    }
  ]
};

export const DEFAULT_ALERTING_QUERY: Partial<WildGraphQLMainQuery> = {
  queryText: `query BatteryVoltage($from: Long!, $to: Long!) {
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
      timePath: "dateMillis"
    }
  ]
};

export const DEFAULT_ANNOTATION_QUERY: Partial<WildGraphQLAnnotationQuery> = {
  queryText: `query BatteryVoltage($from: Long!, $to: Long!) {
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
      timePath: "dateMillis"
    }
  ]
};
