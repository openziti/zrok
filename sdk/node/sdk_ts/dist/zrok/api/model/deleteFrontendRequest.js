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

// src/zrok/api/model/deleteFrontendRequest.ts
var deleteFrontendRequest_exports = {};
__export(deleteFrontendRequest_exports, {
  DeleteFrontendRequest: () => DeleteFrontendRequest
});
module.exports = __toCommonJS(deleteFrontendRequest_exports);
var _DeleteFrontendRequest = class _DeleteFrontendRequest {
  static getAttributeTypeMap() {
    return _DeleteFrontendRequest.attributeTypeMap;
  }
};
_DeleteFrontendRequest.discriminator = void 0;
_DeleteFrontendRequest.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  }
];
var DeleteFrontendRequest = _DeleteFrontendRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  DeleteFrontendRequest
});
//# sourceMappingURL=deleteFrontendRequest.js.map