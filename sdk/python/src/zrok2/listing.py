from zrok_api.api import MetadataApi
from zrok2.environment.root import Root
import zrok2.model as model


def list_shares(root: Root, **filters) -> list[model.ShareDetail]:
    """List shares with optional filters.

    Analogous to ``zrok2 list shares``.

    Args:
        root: The Root environment context.
        **filters: Optional keyword filters passed to the API:
            env_zid, share_mode, backend_mode, share_token, target,
            permission_mode, has_activity, idle, activity_duration,
            created_after, created_before, updated_after, updated_before.

    Returns:
        List of ShareDetail objects.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        metadata_api = MetadataApi(zrok)
        res = metadata_api.list_shares(**filters)
    except Exception as e:
        raise Exception("unable to list shares", e)

    shares = res.shares or []
    return [
        model.ShareDetail(
            Token=s.share_token or "",
            ZId=s.z_id or "",
            EnvZId=s.env_zid or "",
            ShareMode=s.share_mode or "",
            BackendMode=s.backend_mode or "",
            FrontendEndpoints=s.frontend_endpoints or [],
            Target=s.target or "",
            Limited=s.limited or False,
            CreatedAt=s.created_at or 0,
            UpdatedAt=s.updated_at or 0,
        )
        for s in shares
    ]


def list_accesses(root: Root, **filters) -> list[model.AccessDetail]:
    """List accesses with optional filters.

    Analogous to ``zrok2 list accesses``.

    Args:
        root: The Root environment context.
        **filters: Optional keyword filters passed to the API:
            env_zid, share_token, bind_address, description,
            created_after, created_before, updated_after, updated_before.

    Returns:
        List of AccessDetail objects.
    """
    if not root.IsEnabled():
        raise Exception("environment is not enabled; enable with 'zrok enable' first!")

    try:
        zrok = root.Client()
    except Exception as e:
        raise Exception("error getting zrok client", e)

    try:
        metadata_api = MetadataApi(zrok)
        res = metadata_api.list_accesses(**filters)
    except Exception as e:
        raise Exception("unable to list accesses", e)

    accesses = res.accesses or []
    return [
        model.AccessDetail(
            Id=a.id or 0,
            FrontendToken=a.frontend_token or "",
            EnvZId=a.env_zid or "",
            ShareToken=a.share_token or "",
            BackendMode=a.backend_mode or "",
            BindAddress=a.bind_address or "",
            Description=a.description or "",
            Limited=a.limited or False,
            CreatedAt=a.created_at or 0,
            UpdatedAt=a.updated_at or 0,
        )
        for a in accesses
    ]
