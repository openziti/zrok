// src/zrok/api/model/updateFrontendRequest.ts
var _UpdateFrontendRequest = class _UpdateFrontendRequest {
  static getAttributeTypeMap() {
    return _UpdateFrontendRequest.attributeTypeMap;
  }
};
_UpdateFrontendRequest.discriminator = void 0;
_UpdateFrontendRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "publicName",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "urlTemplate",
    "type": "string"
  }
];
var UpdateFrontendRequest = _UpdateFrontendRequest;
export {
  UpdateFrontendRequest
};
//# sourceMappingURL=updateFrontendRequest.mjs.map