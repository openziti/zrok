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

// src/zrok/api/model/inviteRequest.ts
var inviteRequest_exports = {};
__export(inviteRequest_exports, {
  InviteRequest: () => InviteRequest
});
module.exports = __toCommonJS(inviteRequest_exports);
var _InviteRequest = class _InviteRequest {
  static getAttributeTypeMap() {
    return _InviteRequest.attributeTypeMap;
  }
};
_InviteRequest.discriminator = void 0;
_InviteRequest.attributeTypeMap = [
  {
    "name": "email",
    "baseName": "email",
    "type": "string"
  },
  {
    "name": "token",
    "baseName": "token",
    "type": "string"
  }
];
var InviteRequest = _InviteRequest;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  InviteRequest
});
//# sourceMappingURL=inviteRequest.js.map