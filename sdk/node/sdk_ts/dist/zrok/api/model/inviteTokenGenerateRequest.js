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

// src/zrok/api/model/inviteTokenGenerateRequest.ts
var inviteTokenGenerateRequest_exports = {};
__export(inviteTokenGenerateRequest_exports, {
  InviteTokenGenerateRequest: () => InviteTokenGenerateRequest
});
module.exports = __toCommonJS(inviteTokenGenerateRequest_exports);
var _InviteTokenGenerateRequest = class _InviteTokenGenerateRequest {
  static getAttributeTypeMap() {
    return _InviteTokenGenerateRequest.attributeTypeMap;
  }
};
_InviteTokenGenerateRequest.discriminator = void 0;
_InviteTokenGenerateRequest.attributeTypeMap = [
  {
    "name": "tokens",
    "baseName": "tokens",
    "type": "Array<string>"
  }
];
var InviteTokenGenerateRequest = _InviteTokenGenerateRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  InviteTokenGenerateRequest
});
//# sourceMappingURL=inviteTokenGenerateRequest.js.map