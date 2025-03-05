from pathlib import Path
import os


def rootDir() -> str:
    home = str(Path.home())
    return os.path.join(home, ".zrok")


def metadataFile() -> str:
    zrd = rootDir()
    return os.path.join(zrd, "metadata.json")


def configFile() -> str:
    zrd = rootDir()
    return os.path.join(zrd, "config.json")


def environmentFile() -> str:
    zrd = rootDir()
    return os.path.join(zrd, "environment.json")


def identitiesDir() -> str:
    zrd = rootDir()
    return os.path.join(zrd, "identities")


def identityFile(name: str) -> str:
    idd = identitiesDir()
    return os.path.join(idd, name + ".json")
