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

// src/zrok/api/model/resetPasswordRequestRequest.ts
var resetPasswordRequestRequest_exports = {};
__export(resetPasswordRequestRequest_exports, {
  ResetPasswordRequestRequest: () => ResetPasswordRequestRequest
});
module.exports = __toCommonJS(resetPasswordRequestRequest_exports);
var _ResetPasswordRequestRequest = class _ResetPasswordRequestRequest {
  static getAttributeTypeMap() {
    return _ResetPasswordRequestRequest.attributeTypeMap;
  }
};
_ResetPasswordRequestRequest.discriminator = void 0;
_ResetPasswordRequestRequest.attributeTypeMap = [
  {
    "name": "emailAddress",
    "baseName": "emailAddress",
    "type": "string"
  }
];
var ResetPasswordRequestRequest = _ResetPasswordRequestRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  ResetPasswordRequestRequest
});
//# sourceMappingURL=resetPasswordRequestRequest.js.map