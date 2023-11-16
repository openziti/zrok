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

// src/zrok/api/model/metricsSample.ts
var metricsSample_exports = {};
__export(metricsSample_exports, {
  MetricsSample: () => MetricsSample
});
module.exports = __toCommonJS(metricsSample_exports);
var _MetricsSample = class _MetricsSample {
  static getAttributeTypeMap() {
    return _MetricsSample.attributeTypeMap;
  }
};
_MetricsSample.discriminator = void 0;
_MetricsSample.attributeTypeMap = [
  {
    "name": "rx",
    "baseName": "rx",
    "type": "number"
  },
  {
    "name": "tx",
    "baseName": "tx",
    "type": "number"
  },
  {
    "name": "timestamp",
    "baseName": "timestamp",
    "type": "number"
  }
];
var MetricsSample = _MetricsSample;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  MetricsSample
});
//# sourceMappingURL=metricsSample.js.map