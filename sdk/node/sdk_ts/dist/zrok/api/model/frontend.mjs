// src/zrok/api/model/frontend.ts
var _Frontend = class _Frontend {
  static getAttributeTypeMap() {
    return _Frontend.attributeTypeMap;
  }
};
_Frontend.discriminator = void 0;
_Frontend.attributeTypeMap = [
  {
    "name": "id",
    "baseName": "id",
    "type": "number"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
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
var Frontend = _Frontend;
export {
  Frontend
};
//# sourceMappingURL=frontend.mjs.map