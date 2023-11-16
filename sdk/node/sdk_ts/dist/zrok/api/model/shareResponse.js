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

// src/zrok/api/model/shareResponse.ts
var shareResponse_exports = {};
__export(shareResponse_exports, {
  ShareResponse: () => ShareResponse
});
module.exports = __toCommonJS(shareResponse_exports);
var _ShareResponse = class _ShareResponse {
  static getAttributeTypeMap() {
    return _ShareResponse.attributeTypeMap;
  }
};
_ShareResponse.discriminator = void 0;
_ShareResponse.attributeTypeMap = [
  {
    "name": "frontendProxyEndpoints",
    "baseName": "frontendProxyEndpoints",
    "type": "Array<string>"
  },
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  }
];
var ShareResponse = _ShareResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  ShareResponse
});
//# sourceMappingURL=shareResponse.js.map