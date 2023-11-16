// src/zrok/api/model/share.ts
var _Share = class _Share {
  static getAttributeTypeMap() {
    return _Share.attributeTypeMap;
  }
};
_Share.discriminator = void 0;
_Share.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "shareMode",
    "baseName": "shareMode",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  },
  {
    "name": "frontendSelection",
    "baseName": "frontendSelection",
    "type": "string"
  },
  {
    "name": "frontendEndpoint",
    "baseName": "frontendEndpoint",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
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
var Share = _Share;
export {
  Share
};
//# sourceMappingURL=share.mjs.map