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

// src/zrok/api/model/unshareRequest.ts
var unshareRequest_exports = {};
__export(unshareRequest_exports, {
  UnshareRequest: () => UnshareRequest
});
module.exports = __toCommonJS(unshareRequest_exports);
var _UnshareRequest = class _UnshareRequest {
  static getAttributeTypeMap() {
    return _UnshareRequest.attributeTypeMap;
  }
};
_UnshareRequest.discriminator = void 0;
_UnshareRequest.attributeTypeMap = [
  {
    "name": "envZId",
    "baseName": "envZId",
    "type": "string"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "reserved",
    "baseName": "reserved",
    "type": "boolean"
  }
];
var UnshareRequest = _UnshareRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  UnshareRequest
});
//# sourceMappingURL=unshareRequest.js.map