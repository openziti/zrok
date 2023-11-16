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

// src/zrok/api/model/createIdentity201Response.ts
var createIdentity201Response_exports = {};
__export(createIdentity201Response_exports, {
  CreateIdentity201Response: () => CreateIdentity201Response
});
module.exports = __toCommonJS(createIdentity201Response_exports);
var _CreateIdentity201Response = class _CreateIdentity201Response {
  static getAttributeTypeMap() {
    return _CreateIdentity201Response.attributeTypeMap;
  }
};
_CreateIdentity201Response.discriminator = void 0;
_CreateIdentity201Response.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  },
  {
    "name": "cfg",
    "baseName": "cfg",
    "type": "string"
  }
];
var CreateIdentity201Response = _CreateIdentity201Response;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  CreateIdentity201Response
});
//# sourceMappingURL=createIdentity201Response.js.map