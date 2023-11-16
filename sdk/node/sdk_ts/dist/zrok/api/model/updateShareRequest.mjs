// src/zrok/api/model/updateShareRequest.ts
var _UpdateShareRequest = class _UpdateShareRequest {
  static getAttributeTypeMap() {
    return _UpdateShareRequest.attributeTypeMap;
  }
};
_UpdateShareRequest.discriminator = void 0;
_UpdateShareRequest.attributeTypeMap = [
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  }
];
var UpdateShareRequest = _UpdateShareRequest;
export {
  UpdateShareRequest
};
//# sourceMappingURL=updateShareRequest.mjs.map