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

// src/zrok/api/model/environment.ts
var environment_exports = {};
__export(environment_exports, {
  Environment: () => Environment
});
module.exports = __toCommonJS(environment_exports);
var _Environment = class _Environment {
  static getAttributeTypeMap() {
    return _Environment.attributeTypeMap;
  }
};
_Environment.discriminator = void 0;
_Environment.attributeTypeMap = [
  {
    "name": "description",
    "baseName": "description",
    "type": "string"
  },
  {
    "name": "host",
    "baseName": "host",
    "type": "string"
  },
  {
    "name": "address",
    "baseName": "address",
    "type": "string"
  },
  {
    "name": "zId",
    "baseName": "zId",
    "type": "string"
  },
  {
    "name": "activity",
    "baseName": "activity",
    "type": "Array<SparkDataSample>"
  },
  {
    "name": "limited",
    "baseName": "limited",
    "type": "boolean"
  },
  {
    "name": "createdAt",
    "baseName": "createdAt",
    "type": "number"
  },
  {
    "name": "updatedAt",
    "baseName": "updatedAt",
    "type": "number"
  }
];
var Environment = _Environment;
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  Environment
});
//# sourceMappingURL=environment.js.map