// src/zrok/api/model/principal.ts
var _Principal = class _Principal {
  static getAttributeTypeMap() {
    return _Principal.attributeTypeMap;
  }
};
_Principal.discriminator = void 0;
_Principal.attributeTypeMap = [
  {
    "name": "id",
    "baseName": "id",
    "type": "number"
  },
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "limitless",
    "baseName": "limitless",
    "type": "boolean"
  },
  {
    "name": "admin",
    "baseName": "admin",
    "type": "boolean"
  }
];
var Principal = _Principal;
export {
  Principal
};
//# sourceMappingURL=principal.mjs.map