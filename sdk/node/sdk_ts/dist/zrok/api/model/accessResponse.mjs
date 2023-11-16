// src/zrok/api/model/accessResponse.ts
var _AccessResponse = class _AccessResponse {
  static getAttributeTypeMap() {
    return _AccessResponse.attributeTypeMap;
  }
};
_AccessResponse.discriminator = void 0;
_AccessResponse.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  }
];
var AccessResponse = _AccessResponse;
export {
  AccessResponse
};
//# sourceMappingURL=accessResponse.mjs.map