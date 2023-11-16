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

// src/zrok/api/model/registerResponse.ts
var registerResponse_exports = {};
__export(registerResponse_exports, {
  RegisterResponse: () => RegisterResponse
});
module.exports = __toCommonJS(registerResponse_exports);
var _RegisterResponse = class _RegisterResponse {
  static getAttributeTypeMap() {
    return _RegisterResponse.attributeTypeMap;
  }
};
_RegisterResponse.discriminator = void 0;
_RegisterResponse.attributeTypeMap = [
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var RegisterResponse = _RegisterResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  RegisterResponse
});
//# sourceMappingURL=registerResponse.js.map