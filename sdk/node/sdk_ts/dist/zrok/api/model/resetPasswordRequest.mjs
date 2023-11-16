// src/zrok/api/model/resetPasswordRequest.ts
var _ResetPasswordRequest = class _ResetPasswordRequest {
  static getAttributeTypeMap() {
    return _ResetPasswordRequest.attributeTypeMap;
  }
};
_ResetPasswordRequest.discriminator = void 0;
_ResetPasswordRequest.attributeTypeMap = [
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
var ResetPasswordRequest = _ResetPasswordRequest;
export {
  ResetPasswordRequest
};
//# sourceMappingURL=resetPasswordRequest.mjs.map