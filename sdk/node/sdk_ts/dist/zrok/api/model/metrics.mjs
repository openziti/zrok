// src/zrok/api/model/metrics.ts
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
export {
  Metrics
};
//# sourceMappingURL=metrics.mjs.map