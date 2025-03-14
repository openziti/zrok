from zrok.environment.root import Root
import openziti
from socket import SOCK_STREAM


def Dialer(shrToken: str, root: Root) -> openziti.zitisock.ZitiSocket:
    openziti.load(root.ZitiIdentityNamed(root.EnvironmentIdentityName()))
    client = openziti.socket(type=SOCK_STREAM)
    client.connect((shrToken, 1))
    return client
