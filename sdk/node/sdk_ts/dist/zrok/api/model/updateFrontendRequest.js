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

// src/zrok/api/model/updateFrontendRequest.ts
var updateFrontendRequest_exports = {};
__export(updateFrontendRequest_exports, {
  UpdateFrontendRequest: () => UpdateFrontendRequest
});
module.exports = __toCommonJS(updateFrontendRequest_exports);
var _UpdateFrontendRequest = class _UpdateFrontendRequest {
  static getAttributeTypeMap() {
    return _UpdateFrontendRequest.attributeTypeMap;
  }
};
_UpdateFrontendRequest.discriminator = void 0;
_UpdateFrontendRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "publicName",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "urlTemplate",
    "type": "string"
  }
];
var UpdateFrontendRequest = _UpdateFrontendRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  UpdateFrontendRequest
});
//# sourceMappingURL=updateFrontendRequest.js.map