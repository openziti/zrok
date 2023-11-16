// src/zrok/api/model/metricsSample.ts
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
export {
  MetricsSample
};
//# sourceMappingURL=metricsSample.mjs.map