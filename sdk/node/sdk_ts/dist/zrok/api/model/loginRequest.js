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

// src/zrok/api/model/loginRequest.ts
var loginRequest_exports = {};
__export(loginRequest_exports, {
  LoginRequest: () => LoginRequest
});
module.exports = __toCommonJS(loginRequest_exports);
var _LoginRequest = class _LoginRequest {
  static getAttributeTypeMap() {
    return _LoginRequest.attributeTypeMap;
  }
};
_LoginRequest.discriminator = void 0;
_LoginRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "password",
    "baseName": "password",
    "type": "string"
  }
];
var LoginRequest = _LoginRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  LoginRequest
});
//# sourceMappingURL=loginRequest.js.map