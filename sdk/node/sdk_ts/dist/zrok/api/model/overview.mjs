// src/zrok/api/model/overview.ts
var _Overview = class _Overview {
  static getAttributeTypeMap() {
    return _Overview.attributeTypeMap;
  }
};
_Overview.discriminator = void 0;
_Overview.attributeTypeMap = [
  {
    "name": "accountLimited",
    "baseName": "accountLimited",
    "type": "boolean"
  },
  {
    "name": "environments",
    "baseName": "environments",
    "type": "Array<EnvironmentAndResources>"
  }
];
var Overview = _Overview;
export {
  Overview
};
//# sourceMappingURL=overview.mjs.map