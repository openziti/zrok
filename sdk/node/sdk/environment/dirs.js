import os from "os";
import path from "node:path"

function rootDir() {
    home = os.homedir()
    return path.join(home, ".zrok")
}

function metadataFile() {
    zrd = rootDir()
    return path.join(zrd, "metadata.json")
}

function configFile() {
    zrd = rootDir()
    return path.join(zrd, "config.json")
}
function environmentFile() {
    zrd = rootDir()
    return path.join(zrd, "environment.json")
}

function identitiesDir() {
    zrd = rootDir()
    return path.join(zrd, "identities")
}

function identityFile(name) {
    idd = identitiesDir()
    return path.join(idd, name + ".json")
}