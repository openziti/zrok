// src/zrok/api/model/loginRequest.ts
var _LoginRequest = class _LoginRequest {
  static getAttributeTypeMap() {
    return _LoginRequest.attributeTypeMap;
  }
};
_LoginRequest.discriminator = void 0;
_LoginRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var LoginRequest = _LoginRequest;
export {
  LoginRequest
};
//# sourceMappingURL=loginRequest.mjs.map