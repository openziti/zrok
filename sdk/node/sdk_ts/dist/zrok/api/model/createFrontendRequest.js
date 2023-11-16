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

// src/zrok/api/model/createFrontendRequest.ts
var createFrontendRequest_exports = {};
__export(createFrontendRequest_exports, {
  CreateFrontendRequest: () => CreateFrontendRequest
});
module.exports = __toCommonJS(createFrontendRequest_exports);
var _CreateFrontendRequest = class _CreateFrontendRequest {
  static getAttributeTypeMap() {
    return _CreateFrontendRequest.attributeTypeMap;
  }
};
_CreateFrontendRequest.discriminator = void 0;
_CreateFrontendRequest.attributeTypeMap = [
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "urlTemplate",
    "baseName": "url_template",
    "type": "string"
  },
  {
    "name": "publicName",
    "baseName": "public_name",
    "type": "string"
  }
];
var CreateFrontendRequest = _CreateFrontendRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  CreateFrontendRequest
});
//# sourceMappingURL=createFrontendRequest.js.map