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

// src/zrok/api/model/createIdentityRequest.ts
var createIdentityRequest_exports = {};
__export(createIdentityRequest_exports, {
  CreateIdentityRequest: () => CreateIdentityRequest
});
module.exports = __toCommonJS(createIdentityRequest_exports);
var _CreateIdentityRequest = class _CreateIdentityRequest {
  static getAttributeTypeMap() {
    return _CreateIdentityRequest.attributeTypeMap;
  }
};
_CreateIdentityRequest.discriminator = void 0;
_CreateIdentityRequest.attributeTypeMap = [
  {
    "name": "name",
    "baseName": "name",
    "type": "string"
  }
];
var CreateIdentityRequest = _CreateIdentityRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  CreateIdentityRequest
});
//# sourceMappingURL=createIdentityRequest.js.map