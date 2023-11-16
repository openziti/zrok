// src/zrok/api/model/environmentAndResources.ts
var _EnvironmentAndResources = class _EnvironmentAndResources {
  static getAttributeTypeMap() {
    return _EnvironmentAndResources.attributeTypeMap;
  }
};
_EnvironmentAndResources.discriminator = void 0;
_EnvironmentAndResources.attributeTypeMap = [
  {
    "name": "environment",
    "baseName": "environment",
    "type": "Environment"
  },
  {
    "name": "frontends",
    "baseName": "frontends",
    "type": "Array<Frontend>"
  },
  {
    "name": "shares",
    "baseName": "shares",
    "type": "Array<Share>"
  }
];
var EnvironmentAndResources = _EnvironmentAndResources;
export {
  EnvironmentAndResources
};
//# sourceMappingURL=environmentAndResources.mjs.map