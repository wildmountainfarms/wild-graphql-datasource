import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

type VariablesType = string | Record<string, any>;


interface WildGraphQLCommonQuery extends DataQuery {
  queryText: string;
  /** The operation name if explicitly set. Note that an empty string should be treated the same way as an undefined value, although storing an undefined value is preferred.*/
  operationName?: string;
  variables?: VariablesType;
}

export function getQueryVariablesAsJsonString(query: WildGraphQLCommonQuery): string {
  const variables = query.variables;
  if (variables === undefined) {
    return "";
  }
  if (typeof variables === 'string') {
    return variables;
  }
  return JSON.stringify(variables); // TODO consider if we want to prettify this JSON
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


export const DEFAULT_QUERY: Partial<WildGraphQLMainQuery> = {
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
};

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
