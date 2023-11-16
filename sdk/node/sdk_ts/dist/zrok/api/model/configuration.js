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

// src/zrok/api/model/configuration.ts
var configuration_exports = {};
__export(configuration_exports, {
  Configuration: () => Configuration
});
module.exports = __toCommonJS(configuration_exports);
var _Configuration = class _Configuration {
  static getAttributeTypeMap() {
    return _Configuration.attributeTypeMap;
  }
};
_Configuration.discriminator = void 0;
_Configuration.attributeTypeMap = [
  {
    "name": "version",
    "baseName": "version",
    "type": "string"
  },
  {
    "name": "touLink",
    "baseName": "touLink",
    "type": "string"
  },
  {
    "name": "invitesOpen",
    "baseName": "invitesOpen",
    "type": "boolean"
  },
  {
    "name": "requiresInviteToken",
    "baseName": "requiresInviteToken",
    "type": "boolean"
  },
  {
    "name": "inviteTokenContact",
    "baseName": "inviteTokenContact",
    "type": "string"
  },
  {
    "name": "passwordRequirements",
    "baseName": "passwordRequirements",
    "type": "PasswordRequirements"
  }
];
var Configuration = _Configuration;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Configuration
});
//# sourceMappingURL=configuration.js.map