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

// src/zrok/api/model/disableRequest.ts
var disableRequest_exports = {};
__export(disableRequest_exports, {
  DisableRequest: () => DisableRequest
});
module.exports = __toCommonJS(disableRequest_exports);
var _DisableRequest = class _DisableRequest {
  static getAttributeTypeMap() {
    return _DisableRequest.attributeTypeMap;
  }
};
_DisableRequest.discriminator = void 0;
_DisableRequest.attributeTypeMap = [
  {
    "name": "identity",
    "baseName": "identity",
    "type": "string"
  }
];
var DisableRequest = _DisableRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  DisableRequest
});
//# sourceMappingURL=disableRequest.js.map