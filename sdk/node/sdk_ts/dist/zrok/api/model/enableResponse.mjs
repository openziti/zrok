// src/zrok/api/model/enableResponse.ts
var _EnableResponse = class _EnableResponse {
  static getAttributeTypeMap() {
    return _EnableResponse.attributeTypeMap;
  }
};
_EnableResponse.discriminator = void 0;
_EnableResponse.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  },
  {
    "name": "cfg",
    "baseName": "cfg",
    "type": "string"
  }
];
var EnableResponse = _EnableResponse;
export {
  EnableResponse
};
//# sourceMappingURL=enableResponse.mjs.map