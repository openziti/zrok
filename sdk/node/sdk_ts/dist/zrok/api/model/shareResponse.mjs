// src/zrok/api/model/shareResponse.ts
var _ShareResponse = class _ShareResponse {
  static getAttributeTypeMap() {
    return _ShareResponse.attributeTypeMap;
  }
};
_ShareResponse.discriminator = void 0;
_ShareResponse.attributeTypeMap = [
  {
    "name": "frontendProxyEndpoints",
    "baseName": "frontendProxyEndpoints",
    "type": "Array<string>"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var ShareResponse = _ShareResponse;
export {
  ShareResponse
};
//# sourceMappingURL=shareResponse.mjs.map