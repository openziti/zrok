// src/zrok/api/model/authUser.ts
var _AuthUser = class _AuthUser {
  static getAttributeTypeMap() {
    return _AuthUser.attributeTypeMap;
  }
};
_AuthUser.discriminator = void 0;
_AuthUser.attributeTypeMap = [
  {
    "name": "username",
    "baseName": "username",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var AuthUser = _AuthUser;
export {
  AuthUser
};
//# sourceMappingURL=authUser.mjs.map