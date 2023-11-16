// src/zrok/api/model/enableRequest.ts
var _EnableRequest = class _EnableRequest {
  static getAttributeTypeMap() {
    return _EnableRequest.attributeTypeMap;
  }
};
_EnableRequest.discriminator = void 0;
_EnableRequest.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  }
];
var EnableRequest = _EnableRequest;
export {
  EnableRequest
};
//# sourceMappingURL=enableRequest.mjs.map