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

// src/zrok/api/model/accessRequest.ts
var accessRequest_exports = {};
__export(accessRequest_exports, {
  AccessRequest: () => AccessRequest
});
module.exports = __toCommonJS(accessRequest_exports);
var _AccessRequest = class _AccessRequest {
  static getAttributeTypeMap() {
    return _AccessRequest.attributeTypeMap;
  }
};
_AccessRequest.discriminator = void 0;
_AccessRequest.attributeTypeMap = [
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
var AccessRequest = _AccessRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  AccessRequest
});
//# sourceMappingURL=accessRequest.js.map