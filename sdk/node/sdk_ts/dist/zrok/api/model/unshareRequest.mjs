// src/zrok/api/model/unshareRequest.ts
var _UnshareRequest = class _UnshareRequest {
  static getAttributeTypeMap() {
    return _UnshareRequest.attributeTypeMap;
  }
};
_UnshareRequest.discriminator = void 0;
_UnshareRequest.attributeTypeMap = [
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  }
];
var UnshareRequest = _UnshareRequest;
export {
  UnshareRequest
};
//# sourceMappingURL=unshareRequest.mjs.map