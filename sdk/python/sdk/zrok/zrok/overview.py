from zrok.environment.root import Root
import urllib3


def Overview(root: Root) -> str:
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    http = urllib3.PoolManager()
    apiEndpoint = root.ApiEndpoint().endpoint
    try:
        response = http.request(
            'GET',
            apiEndpoint + "/api/v1/overview",
            headers={
                "X-TOKEN": root.env.Token
            })
    except Exception as e:
        raise Exception("unable to get account overview", e)
    return response.data.decode('utf-8')
