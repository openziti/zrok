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

// src/zrok/api/model/updateShareRequest.ts
var updateShareRequest_exports = {};
__export(updateShareRequest_exports, {
  UpdateShareRequest: () => UpdateShareRequest
});
module.exports = __toCommonJS(updateShareRequest_exports);
var _UpdateShareRequest = class _UpdateShareRequest {
  static getAttributeTypeMap() {
    return _UpdateShareRequest.attributeTypeMap;
  }
};
_UpdateShareRequest.discriminator = void 0;
_UpdateShareRequest.attributeTypeMap = [
  {
    "name": "shrToken",
    "baseName": "shrToken",
    "type": "string"
  },
  {
    "name": "backendProxyEndpoint",
    "baseName": "backendProxyEndpoint",
    "type": "string"
  }
];
var UpdateShareRequest = _UpdateShareRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  UpdateShareRequest
});
//# sourceMappingURL=updateShareRequest.js.map