import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import {Share} from "./api";
import {useEffect, useState} from "react";
import PropertyTable from "./PropertyTable.tsx";
import useApiConsoleStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import ReleaseShareModal from "./ReleaseShareModal.tsx";
import {getMetadataApi} from "./model/api.ts";
import ClipboardText from "./ClipboardText.tsx";

interface SharePanelProps {
    share: Node;
}

const SharePanel = ({ share }: SharePanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const [detail, setDetail] = useState<Share>(null);
    const [releaseShareOpen, setReleaseShareOpen] = useState(false);

    const openReleaseShare = () => {
        setReleaseShareOpen(true);
    }

    const closeReleaseShare = () => {
        setReleaseShareOpen(false);
    }

    const customProperties = {
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString(),
        frontendEndpoint: row => <>
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <a href={row.value} target="_">{row.value}</a>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ClipboardText text={row.value} />
                </Grid2>
            </Grid2>
        </>,
        reserved: row => row.value ? "reserved" : "ephemeral",
        token: row => <>
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{row.value}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ClipboardText text={row.value} />
                </Grid2>
            </Grid2>
        </>
    }

    const labels = {
        backendProxyEndpoint: "Target",
        createdAt: "Created",
        reserved: "Reservation",
        updatedAt: "Updated"
    }

    useEffect(() => {
        getMetadataApi(user).getShareDetail({ shrToken: share.data!.shrToken! as string })
            .then(d => {
                delete d.activity;
                delete d.limited;
                delete d.zId;
                if(d.shareMode === "private") {
                    delete d.frontendEndpoint;
                    delete d.frontendSelection;
                }
                setDetail(d);
            })
            .catch(e => {
                console.log("SharePanel", e);
            })
    }, [share]);

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 0 }} alignItems="center">
                    <Grid2 display="flex"><ShareIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h4">{String(share.data.label)}</Grid2>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2 }} alignItems="center">
                    <h5 style={{ margin: 0 }}>A {detail ? detail.shareMode : ''}{detail && detail.reserved ? ', reserved ' : ''} share with the token <code>{share.id}</code></h5>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Release Share">
                        <Button variant="contained" color="error" onClick={openReleaseShare}><DeleteIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                    </Grid2>
                </Grid2>
            </Typography>
            <ReleaseShareModal close={closeReleaseShare} isOpen={releaseShareOpen} user={user} share={share} detail={detail} />
        </>
    );
}

export default SharePanel;
