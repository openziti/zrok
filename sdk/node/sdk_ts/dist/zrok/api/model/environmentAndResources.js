"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/zrok/api/model/environmentAndResources.ts
var environmentAndResources_exports = {};
__export(environmentAndResources_exports, {
  EnvironmentAndResources: () => EnvironmentAndResources
});
module.exports = __toCommonJS(environmentAndResources_exports);
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  EnvironmentAndResources
});
//# sourceMappingURL=environmentAndResources.js.map