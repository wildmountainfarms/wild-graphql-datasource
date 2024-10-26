import {getTemplateSrv, TemplateSrv} from "@grafana/runtime";
import {ScopedVars} from "@grafana/data";

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
const AUTO_POPULATED_VARIABLES: Record<string, (templateSrv: TemplateSrv) => any> = {
  "from": templateSrv => Number(templateSrv.replace("$__from")),
  "to": templateSrv => Number(templateSrv.replace("$__to")),
  // While interval_ms can be obtained via $__interval_ms, but only as a ScopedVar, which we don't have easy access to inside a fetcher
};

/**
 * This should only be used for client-side only queries, such as the Execute button.
 * Remember that this implementation is not meant to be perfect, but an approximation of how the backend functions
 */
export function getInterpolatedAutoPopulatedVariables(templateSrv: TemplateSrv): Record<string, any> {
  const variables: any = {};
  for (const variableName in AUTO_POPULATED_VARIABLES) {
    const func = AUTO_POPULATED_VARIABLES[variableName];
    const result = func(templateSrv);
    if (isNaN(result)) {
      console.error("Could not add interpolation for variable: " + variableName + ". Will not pass as a variable.");
    } else {
      variables[variableName] = result;
    }
  }

  return variables;
}



function doInterpolate(object: any, templateSrv: TemplateSrv, scopedVars?: ScopedVars): any {
  switch (typeof object) {
    case 'string':
      return templateSrv.replace(object, scopedVars)
    case 'object':
      if (Array.isArray(object)) {
        return object.map(value => doInterpolate(value, templateSrv, scopedVars));
      } else {
        const newObject: any = {};
        for (const field in object) {
          newObject[field] = doInterpolate(object[field], templateSrv, scopedVars);
        }
        return newObject;
      }
  }
  return object;
}

export function interpolateVariables(variables: Record<string, any>, templateSrv: TemplateSrv, scopedVars?: ScopedVars): Record<string, any> {
  return doInterpolate(variables, templateSrv, scopedVars);
}

