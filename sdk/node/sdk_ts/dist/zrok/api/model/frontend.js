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

// src/zrok/api/model/frontend.ts
var frontend_exports = {};
__export(frontend_exports, {
  Frontend: () => Frontend
});
module.exports = __toCommonJS(frontend_exports);
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
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Frontend
});
//# sourceMappingURL=frontend.js.map