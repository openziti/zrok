import {Node} from "@xyflow/react";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Tooltip, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import {Configuration, MetadataApi, Share, ShareApi} from "./api";
import {useEffect, useRef, useState} from "react";
import PropertyTable from "./PropertyTable.tsx";
import useStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import {modalStyle} from "./styling/theme.ts";

interface ReleaseShareProps {
    close: () => void;
    isOpen: boolean;
    detail: Share;
    action: () => void;
}

const ReleaseShareModal = ({ close, isOpen, detail, action }: ReleaseShareProps) => {
    const [token, setToken] = useState<String>("");
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>();
    checkedRef.current = checked;

    const toggleChecked = (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
    }, [isOpen]);

    useEffect(() => {
        if(detail && detail.token) {
            setToken(detail.token);
        }
    }, [detail]);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Release Share</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Would you like to release the share <code>{token}</code> ?</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm the release of <code>{token}</code></p>} sx={{ mt: 2 }} />
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked} onClick={action}>Release</Button>
                </Grid2>
            </Box>
        </Modal>
    )
}

interface SharePanelProps {
    share: Node;
}

const SharePanel = ({ share }: SharePanelProps) => {
    const user = useStore((state) => state.user);
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
        frontendEndpoint: row => <a href={row.value} target="_">{row.value}</a>,
        reserved: row => row.value ? "reserved" : "ephemeral"
    }

    const labels = {
        backendProxyEndpoint: "Target",
        createdAt: "Created",
        reserved: "Reservation",
        updatedAt: "Updated"
    }

    useEffect(() => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let metadata = new MetadataApi(cfg);
        metadata.getShareDetail({ shrToken: share.data!.shrToken! as string })
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

    const releaseShare = () => {
        if(detail) {
            let cfg = new Configuration({
                headers: {
                    "X-TOKEN": user.token
                }
            });
            let shareApi = new ShareApi(cfg);
            shareApi.unshare({body: {envZId: share.data.envZId as string, shrToken: detail.token, reserved: detail.reserved}})
                .then(d => {
                    setReleaseShareOpen(false);
                })
                .catch(e => {
                    console.log("releaseShare", e);
                });
        }
    }

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Grid2 display="flex"><ShareIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h3">{String(share.data.label)}</Grid2>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Release Environment">
                        <Button variant="contained" color="error" onClick={openReleaseShare}><DeleteIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                    </Grid2>
                </Grid2>
            </Typography>
            <ReleaseShareModal close={closeReleaseShare} isOpen={releaseShareOpen} detail={detail} action={releaseShare} />
        </>
    );
}

export default SharePanel;
