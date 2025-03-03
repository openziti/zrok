import { homedir } from "os";
import { join } from "node:path";

export const rootDir = (): string => {
    return join(homedir(), ".zrok");
}

export const metadataFile = (): string => {
    return join(rootDir(), "metadata.json");
}

export const environmentFile = (): string => {
    return join(rootDir(), "environment.json");
}

export const identitiesDir = (): string => {
    return join(rootDir(), "identities");
}

export const identityFile = (name: string): string => {
    return join(identitiesDir(), name + ".json");
}