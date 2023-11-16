// src/zrok/api/model/publicFrontend.ts
var _PublicFrontend = class _PublicFrontend {
  static getAttributeTypeMap() {
    return _PublicFrontend.attributeTypeMap;
  }
};
_PublicFrontend.discriminator = void 0;
_PublicFrontend.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "urlTemplate",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "publicName",
    "type": "string"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var PublicFrontend = _PublicFrontend;
export {
  PublicFrontend
};
//# sourceMappingURL=publicFrontend.mjs.map