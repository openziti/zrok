// src/zrok/api/model/environment.ts
var _Environment = class _Environment {
  static getAttributeTypeMap() {
    return _Environment.attributeTypeMap;
  }
};
_Environment.discriminator = void 0;
_Environment.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  },
  {
    "name": "address",
    "baseName": "address",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "activity",
    "baseName": "activity",
    "type": "Array<SparkDataSample>"
  },
  {
    "name": "limited",
    "baseName": "limited",
    "type": "boolean"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var Environment = _Environment;
export {
  Environment
};
//# sourceMappingURL=environment.mjs.map