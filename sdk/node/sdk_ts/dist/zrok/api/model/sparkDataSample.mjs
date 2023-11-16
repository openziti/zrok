// src/zrok/api/model/sparkDataSample.ts
var _SparkDataSample = class _SparkDataSample {
  static getAttributeTypeMap() {
    return _SparkDataSample.attributeTypeMap;
  }
};
_SparkDataSample.discriminator = void 0;
_SparkDataSample.attributeTypeMap = [
  {
    "name": "rx",
    "baseName": "rx",
    "type": "number"
  },
  {
    "name": "tx",
    "baseName": "tx",
    "type": "number"
  }
];
var SparkDataSample = _SparkDataSample;
export {
  SparkDataSample
};
//# sourceMappingURL=sparkDataSample.mjs.map