// src/zrok/api/model/createFrontendRequest.ts
var _CreateFrontendRequest = class _CreateFrontendRequest {
  static getAttributeTypeMap() {
    return _CreateFrontendRequest.attributeTypeMap;
  }
};
_CreateFrontendRequest.discriminator = void 0;
_CreateFrontendRequest.attributeTypeMap = [
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "url_template",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "public_name",
    "type": "string"
  }
];
var CreateFrontendRequest = _CreateFrontendRequest;
export {
  CreateFrontendRequest
};
//# sourceMappingURL=createFrontendRequest.mjs.map