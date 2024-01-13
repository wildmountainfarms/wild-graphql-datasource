import {getTemplateSrv} from "@grafana/runtime";

/**
 * This represents variables that are automatically populated.
 * The keys are the variable names, and the values should be interpolated by {@link getTemplateSrv}.
 *
 * The values should only be used when making a query directly from the frontend (without going through the Go backend).
 *
 * This can be used as a reference for which values the backend should replace.
 *
 * NOTE: Do not add variables to this that the backend cannot supply.
 * User provided variables should be interpolated on the frontend when possible,
 * but we want to reduce the dependency on the frontend here specifically because the backend cannot use {@link getTemplateSrv}.
 * Remember THE VALUES HERE ARE NOT USED BY THE BACKEND AND ARE ONLY USED FOR DEBUGGING QUERIES IN THE FRONTEND BY THE RUN BUTTON.
 */
export const AUTO_POPULATED_VARIABLES: Record<string, any> = {
  // TODO pass these values as numbers after interpolating them
  "from": "$__from",
  "to": "$__to",
  "interval_ms": "$__interval_ms",
};

