import {TemplateSrv} from "@grafana/runtime";
import {interpolateVariables} from "./variables";


import {ScopedVars} from '@grafana/data';

function createTemplateSrv(): TemplateSrv {
  return {
    containsTemplate: jest.fn(),
    updateTimeRange: jest.fn(),
    replace: jest.fn().mockImplementation((value) => `replaced:${value}`),
    getVariables: jest.fn()
  };
}

describe("interpolateVariables", () => {
  // Good example here: https://github.com/RedisGrafana/grafana-redis-datasource/blob/33915a452abcd0016447fb8a881575252605c10e/src/datasource/datasource.test.ts#L221
  const scopedVars: ScopedVars | undefined = { keyName: { value: 'key', text: '' } };

  it("Flat interpolations", () => {
    const templateSrv: TemplateSrv = createTemplateSrv();
    const variables = {
      "sourceId": "source_id_value_here"
    };
    const result = interpolateVariables(variables, templateSrv, scopedVars);
    expect(result).toEqual({
      "sourceId": "replaced:source_id_value_here"
    });
    expect(templateSrv.replace).toHaveBeenCalledWith("source_id_value_here", scopedVars)
    expect(templateSrv.replace).toHaveBeenCalledTimes(1)
  });
  it("No interpolations", () => {
    const templateSrv: TemplateSrv = createTemplateSrv();
    const variables = {
      "someValue": 5
    };
    const result = interpolateVariables(variables, templateSrv, scopedVars);
    expect(result).toEqual(variables);
    expect(templateSrv.replace).toHaveBeenCalledTimes(0)
  });
  it("Complex interpolations", () => {
    const templateSrv: TemplateSrv = createTemplateSrv();
    const variables = {
      "someValue": 5,
      "coolObject": {
        "foo": 6,
        "fee": "fi",
        "bar": {
          "boo": "asdf"
        }
      }
    };
    const result = interpolateVariables(variables, templateSrv, scopedVars);
    expect(result).toEqual({
      "someValue": 5,
      "coolObject": {
        "foo": 6,
        "fee": "replaced:fi",
        "bar": {
          "boo": "replaced:asdf"
        }
      }
    });
    expect(templateSrv.replace).toHaveBeenCalledTimes(2)
  });
  it("Interpolation with array", () => {
    const templateSrv: TemplateSrv = createTemplateSrv();
    const variables = {
      "prices": ["$9.99", 5.0]
    };
    const result = interpolateVariables(variables, templateSrv, scopedVars);
    expect(result).toEqual({
      "prices": ["replaced:$9.99", 5.0]
    });
  });
});
