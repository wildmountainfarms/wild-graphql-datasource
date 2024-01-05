import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

interface WildGraphQLCommonQuery extends DataQuery {
  queryText: string;
  operationName: string;
}

/**
 * This interface represents the main type of query, which is used for most queries in Grafana
 */
export interface WildGraphQLMainQuery extends WildGraphQLCommonQuery {
}

export const DEFAULT_QUERY: Partial<WildGraphQLMainQuery> = {
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
