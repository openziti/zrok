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

// src/zrok/api/model/unaccessRequest.ts
var unaccessRequest_exports = {};
__export(unaccessRequest_exports, {
  UnaccessRequest: () => UnaccessRequest
});
module.exports = __toCommonJS(unaccessRequest_exports);
var _UnaccessRequest = class _UnaccessRequest {
  static getAttributeTypeMap() {
    return _UnaccessRequest.attributeTypeMap;
  }
};
_UnaccessRequest.discriminator = void 0;
_UnaccessRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var UnaccessRequest = _UnaccessRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  UnaccessRequest
});
//# sourceMappingURL=unaccessRequest.js.map