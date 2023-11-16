"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/zrok/api/model/metrics.ts
var metrics_exports = {};
__export(metrics_exports, {
  Metrics: () => Metrics
});
module.exports = __toCommonJS(metrics_exports);
var _Metrics = class _Metrics {
  static getAttributeTypeMap() {
    return _Metrics.attributeTypeMap;
  }
};
_Metrics.discriminator = void 0;
_Metrics.attributeTypeMap = [
  {
    "name": "scope",
    "baseName": "scope",
    "type": "string"
  },
  {
    "name": "id",
    "baseName": "id",
    "type": "string"
  },
  {
    "name": "period",
    "baseName": "period",
    "type": "number"
  },
  {
    "name": "samples",
    "baseName": "samples",
    "type": "Array<MetricsSample>"
  }
];
var Metrics = _Metrics;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Metrics
});
//# sourceMappingURL=metrics.js.map