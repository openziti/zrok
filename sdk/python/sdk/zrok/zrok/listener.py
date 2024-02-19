import openziti
from zrok.environment.root import Root


class Listener():
    shrToken: str
    root: Root
    __server: openziti.zitisock.ZitiSocket

    def __init__(self, shrToken: str, root: Root):
        self.shrToken = shrToken
        self.root = root
        ztx = openziti.load(
            self.root.ZitiIdentityNamed(
                self.root.EnvironmentIdentityName()))
        self.__server = ztx.bind(self.shrToken)

    def __enter__(self) -> openziti.zitisock.ZitiSocket:
        self.listen()
        return self.__server

    def __exit__(self, exception_type, exception_value, exception_traceback):
        self.close()

    def listen(self):
        self.__server.listen()

    def close(self):
        self.__server.close()
