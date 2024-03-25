import { homedir } from "os"
import { join } from "node:path"

export function rootDir(): string {
    return join(homedir(), ".zrok")
}

export function metadataFile(): string {
    return join(rootDir(), "metadata.json")
}

export function configFile(): string {
    return join(rootDir(), "config.json")
}

export function environmentFile(): string {
    return join(rootDir(), "environment.json")
}

export function identitiesDir(): string {
    return join(rootDir(), "identities")
}

export function identityFile(name: string): string {
    return join(identitiesDir(), name + ".json")
}