// src/zrok/api/model/registerRequest.ts
var _RegisterRequest = class _RegisterRequest {
  static getAttributeTypeMap() {
    return _RegisterRequest.attributeTypeMap;
  }
};
_RegisterRequest.discriminator = void 0;
_RegisterRequest.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var RegisterRequest = _RegisterRequest;
export {
  RegisterRequest
};
//# sourceMappingURL=registerRequest.mjs.map