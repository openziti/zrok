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

// src/zrok/api/model/share.ts
var share_exports = {};
__export(share_exports, {
  Share: () => Share
});
module.exports = __toCommonJS(share_exports);
var _Share = class _Share {
  static getAttributeTypeMap() {
    return _Share.attributeTypeMap;
  }
};
_Share.discriminator = void 0;
_Share.attributeTypeMap = [
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
    "name": "shareMode",
    "baseName": "shareMode",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  },
  {
    "name": "frontendSelection",
    "baseName": "frontendSelection",
    "type": "string"
  },
  {
    "name": "frontendEndpoint",
    "baseName": "frontendEndpoint",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  },
  {
    "name": "activity",
    "baseName": "activity",
    "type": "Array<SparkDataSample>"
  },
  {
    "name": "limited",
    "baseName": "limited",
    "type": "boolean"
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
var Share = _Share;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Share
});
//# sourceMappingURL=share.js.map