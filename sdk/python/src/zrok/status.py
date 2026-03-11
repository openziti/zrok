from zrok.environment.root import Root
import zrok.model as model


def status(root: Root) -> model.Status:
    """Return environment status.

    Analogous to ``zrok2 status``.

    Args:
        root: The Root environment context.

    Returns:
        A Status object with environment state information.
    """
    api_endpoint = root.ApiEndpoint()

    return model.Status(
        Enabled=root.IsEnabled(),
        ApiEndpoint=api_endpoint.endpoint,
        ApiEndpointSource=api_endpoint.frm,
        Token=root.env.Token if root.IsEnabled() else "",
        ZitiIdentity=root.env.ZitiIdentity if root.IsEnabled() else "",
    )
