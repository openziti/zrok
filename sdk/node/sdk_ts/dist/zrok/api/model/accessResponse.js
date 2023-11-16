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

// src/zrok/api/model/accessResponse.ts
var accessResponse_exports = {};
__export(accessResponse_exports, {
  AccessResponse: () => AccessResponse
});
module.exports = __toCommonJS(accessResponse_exports);
var _AccessResponse = class _AccessResponse {
  static getAttributeTypeMap() {
    return _AccessResponse.attributeTypeMap;
  }
};
_AccessResponse.discriminator = void 0;
_AccessResponse.attributeTypeMap = [
  {
    "name": "frontendToken",
    "baseName": "frontendToken",
    "type": "string"
  },
  {
    "name": "backendMode",
    "baseName": "backendMode",
    "type": "string"
  }
];
var AccessResponse = _AccessResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  AccessResponse
});
//# sourceMappingURL=accessResponse.js.map