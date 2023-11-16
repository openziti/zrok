// src/zrok/api/model/unaccessRequest.ts
var _UnaccessRequest = class _UnaccessRequest {
  static getAttributeTypeMap() {
    return _UnaccessRequest.attributeTypeMap;
  }
};
_UnaccessRequest.discriminator = void 0;
_UnaccessRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var UnaccessRequest = _UnaccessRequest;
export {
  UnaccessRequest
};
//# sourceMappingURL=unaccessRequest.mjs.map