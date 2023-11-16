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

// src/zrok/api/model/enableRequest.ts
var enableRequest_exports = {};
__export(enableRequest_exports, {
  EnableRequest: () => EnableRequest
});
module.exports = __toCommonJS(enableRequest_exports);
var _EnableRequest = class _EnableRequest {
  static getAttributeTypeMap() {
    return _EnableRequest.attributeTypeMap;
  }
};
_EnableRequest.discriminator = void 0;
_EnableRequest.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  }
];
var EnableRequest = _EnableRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  EnableRequest
});
//# sourceMappingURL=enableRequest.js.map