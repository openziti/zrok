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

// src/zrok/api/model/enableResponse.ts
var enableResponse_exports = {};
__export(enableResponse_exports, {
  EnableResponse: () => EnableResponse
});
module.exports = __toCommonJS(enableResponse_exports);
var _EnableResponse = class _EnableResponse {
  static getAttributeTypeMap() {
    return _EnableResponse.attributeTypeMap;
  }
};
_EnableResponse.discriminator = void 0;
_EnableResponse.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  },
  {
    "name": "cfg",
    "baseName": "cfg",
    "type": "string"
  }
];
var EnableResponse = _EnableResponse;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  EnableResponse
});
//# sourceMappingURL=enableResponse.js.map