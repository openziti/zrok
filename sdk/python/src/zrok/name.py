from zrok_api.api import ShareApi
from zrok_api.models.create_share_name_request import CreateShareNameRequest
from zrok_api.exceptions import ApiException
from zrok.environment.root import Root
import zrok.model as model


_CUSTOM_HEADERS = {
    'Accept': 'application/json, application/zrok.v1+json'
}


def create_name(root: Root, name: str, namespace_token: str = None) -> model.NameEntry:
    """Create a name in a namespace.

    Analogous to ``zrok2 create name <name>``.

    Args:
        root: The Root environment context.
        name: The name to create.
        namespace_token: Optional namespace token. Uses default namespace if omitted.

    Returns:
        A NameEntry with the created name metadata.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    req = CreateShareNameRequest(
        namespace_token=namespace_token,
        name=name,
    )

    try:
        share_api = ShareApi(zrok)
        share_api.create_share_name_with_http_info(body=req, _headers=_CUSTOM_HEADERS)
    except ApiException as e:
        if "Unsupported content type: application/zrok.v1+json" in str(e) and (200 <= e.status <= 299):
            pass
        else:
            raise Exception(f"unable to create name '{name}'", e)
    except Exception as e:
        raise Exception(f"unable to create name '{name}'", e)

    return model.NameEntry(
        NamespaceToken=namespace_token or "",
        Name=name,
    )


def delete_name(root: Root, name: str, namespace_token: str = None) -> None:
    """Delete a name from a namespace.

    Analogous to ``zrok2 delete name <name>``.

    Args:
        root: The Root environment context.
        name: The name to delete.
        namespace_token: Optional namespace token.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    req = CreateShareNameRequest(
        namespace_token=namespace_token,
        name=name,
    )

    try:
        share_api = ShareApi(zrok)
        share_api.delete_share_name_with_http_info(body=req, _headers=_CUSTOM_HEADERS)
    except ApiException as e:
        if "Unsupported content type: application/zrok.v1+json" in str(e) and (200 <= e.status <= 299):
            pass
        else:
            raise Exception(f"unable to delete name '{name}'", e)
    except Exception as e:
        raise Exception(f"unable to delete name '{name}'", e)


def list_names(root: Root, namespace_token: str = None) -> list[model.NameEntry]:
    """List names in a namespace.

    Analogous to ``zrok2 list names``.

    Args:
        root: The Root environment context.
        namespace_token: If provided, list names only in this namespace.

    Returns:
        List of NameEntry objects.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        share_api = ShareApi(zrok)
        if namespace_token:
            res = share_api.list_names_for_namespace(namespace_token=namespace_token)
        else:
            # list_share_namespaces returns namespaces; for all names we iterate
            namespaces = share_api.list_share_namespaces()
            res = []
            for ns in namespaces:
                ns_names = share_api.list_names_for_namespace(namespace_token=ns.namespace_token)
                res.extend(ns_names)
    except Exception as e:
        raise Exception("unable to list names", e)

    return [
        model.NameEntry(
            NamespaceToken=n.namespace_token or "",
            NamespaceName=n.namespace_name or "",
            Name=n.name or "",
            ShareToken=n.share_token or "",
            Reserved=n.reserved or False,
            CreatedAt=n.created_at or 0,
        )
        for n in res
    ]


def list_namespaces(root: Root) -> list[model.Namespace]:
    """List all available namespaces.

    Analogous to ``zrok2 list namespaces``.

    Args:
        root: The Root environment context.

    Returns:
        List of Namespace objects.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        share_api = ShareApi(zrok)
        res = share_api.list_share_namespaces()
    except Exception as e:
        raise Exception("unable to list namespaces", e)

    return [
        model.Namespace(
            NamespaceToken=ns.namespace_token or "",
            Name=ns.name or "",
            Description=ns.description or "",
        )
        for ns in res
    ]
