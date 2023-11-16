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

// src/environment/dirs.ts
var dirs_exports = {};
__export(dirs_exports, {
  configFile: () => configFile,
  environmentFile: () => environmentFile,
  identitiesDir: () => identitiesDir,
  identityFile: () => identityFile,
  metadataFile: () => metadataFile,
  rootDir: () => rootDir
});
module.exports = __toCommonJS(dirs_exports);
var import_os = require("os");
var import_node_path = require("path");
function rootDir() {
  return (0, import_node_path.join)((0, import_os.homedir)(), ".zrok");
}
function metadataFile() {
  return (0, import_node_path.join)(rootDir(), "metadata.json");
}
function configFile() {
  return (0, import_node_path.join)(rootDir(), "config.json");
}
function environmentFile() {
  return (0, import_node_path.join)(rootDir(), "environment.json");
}
function identitiesDir() {
  return (0, import_node_path.join)(rootDir(), "identities");
}
function identityFile(name) {
  return (0, import_node_path.join)(identitiesDir(), name + ".json");
}
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  configFile,
  environmentFile,
  identitiesDir,
  identityFile,
  metadataFile,
  rootDir
});
//# sourceMappingURL=dirs.js.map