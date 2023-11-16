// src/zrok/api/model/accessRequest.ts
var _AccessRequest = class _AccessRequest {
  static getAttributeTypeMap() {
    return _AccessRequest.attributeTypeMap;
  }
};
_AccessRequest.discriminator = void 0;
_AccessRequest.attributeTypeMap = [
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
var AccessRequest = _AccessRequest;
export {
  AccessRequest
};
//# sourceMappingURL=accessRequest.mjs.map