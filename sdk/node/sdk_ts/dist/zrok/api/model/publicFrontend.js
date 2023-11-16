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

// src/zrok/api/model/publicFrontend.ts
var publicFrontend_exports = {};
__export(publicFrontend_exports, {
  PublicFrontend: () => PublicFrontend
});
module.exports = __toCommonJS(publicFrontend_exports);
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  PublicFrontend
});
//# sourceMappingURL=publicFrontend.js.map