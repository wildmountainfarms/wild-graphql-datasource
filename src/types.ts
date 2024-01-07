import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

interface WildGraphQLCommonQuery extends DataQuery {
  queryText: string;
  /** The operation name if explicitly set. Note that an empty string should be treated the same way as an undefined value, although storing an undefined value is preferred.*/
  operationName?: string;
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
