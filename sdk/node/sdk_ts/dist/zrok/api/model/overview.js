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

// src/zrok/api/model/overview.ts
var overview_exports = {};
__export(overview_exports, {
  Overview: () => Overview
});
module.exports = __toCommonJS(overview_exports);
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Overview
});
//# sourceMappingURL=overview.js.map