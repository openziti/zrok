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

// src/zrok/api/model/verifyResponse.ts
var verifyResponse_exports = {};
__export(verifyResponse_exports, {
  VerifyResponse: () => VerifyResponse
});
module.exports = __toCommonJS(verifyResponse_exports);
var _VerifyResponse = class _VerifyResponse {
  static getAttributeTypeMap() {
    return _VerifyResponse.attributeTypeMap;
  }
};
_VerifyResponse.discriminator = void 0;
_VerifyResponse.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  }
];
var VerifyResponse = _VerifyResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  VerifyResponse
});
//# sourceMappingURL=verifyResponse.js.map