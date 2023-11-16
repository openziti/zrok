// src/environment/dirs.ts
import { homedir } from "os";
import { join } from "path";
function rootDir() {
  return join(homedir(), ".zrok");
}
function metadataFile() {
  return join(rootDir(), "metadata.json");
}
function configFile() {
  return join(rootDir(), "config.json");
}
function environmentFile() {
  return join(rootDir(), "environment.json");
}
function identitiesDir() {
  return join(rootDir(), "identities");
}
function identityFile(name) {
  return join(identitiesDir(), name + ".json");
}
export {
  configFile,
  environmentFile,
  identitiesDir,
  identityFile,
  metadataFile,
  rootDir
};
//# sourceMappingURL=dirs.mjs.map