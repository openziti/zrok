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

// src/zrok/api/model/createFrontendResponse.ts
var createFrontendResponse_exports = {};
__export(createFrontendResponse_exports, {
  CreateFrontendResponse: () => CreateFrontendResponse
});
module.exports = __toCommonJS(createFrontendResponse_exports);
var _CreateFrontendResponse = class _CreateFrontendResponse {
  static getAttributeTypeMap() {
    return _CreateFrontendResponse.attributeTypeMap;
  }
};
_CreateFrontendResponse.discriminator = void 0;
_CreateFrontendResponse.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var CreateFrontendResponse = _CreateFrontendResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  CreateFrontendResponse
});
//# sourceMappingURL=createFrontendResponse.js.map