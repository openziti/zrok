import {Node} from "@xyflow/react";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Tooltip, Typography} from "@mui/material";
import AccessIcon from "@mui/icons-material/Lan";
import useStore from "./model/store.ts";
import {useEffect, useRef, useState} from "react";
import {Configuration, Frontend, MetadataApi, ShareApi} from "./api";
import DeleteIcon from "@mui/icons-material/Delete";
import PropertyTable from "./PropertyTable.tsx";
import {modalStyle} from "./styling/theme.ts";

interface ReleaseAccessProps {
    close: () => void;
    isOpen: boolean;
    detail: Frontend;
    action: () => void;
}

const ReleaseAccessModal = ({ close, isOpen, detail, action }: ReleaseAccessProps) => {
    const [feToken, setFeToken] = useState<String>("");
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>(checked);

    const toggleChecked = (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
    }, [isOpen]);

    useEffect(() => {
        if(detail && detail.token) {
            setFeToken(detail.token);
        }
    }, [detail]);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Release Access</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Would you like to release the access <code>{feToken}</code> ?</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm the release of <code>{feToken}</code></p>} sx={{ mt: 2 }} />
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked} onClick={action}>Release</Button>
                </Grid2>
            </Box>
        </Modal>
    );
}

interface AccessPanelProps {
    access: Node;
}

const AccessPanel = ({ access }: AccessPanelProps) => {
    const user = useStore((state) => state.user);
    const [detail, setDetail] = useState<Frontend>(null);
    const [releaseAccessOpen, setReleaseAccessOpen] = useState(false);

    const openReleaseAccess = () => {
        setReleaseAccessOpen(true);
    }

    const closeReleaseAccess = () => {
        setReleaseAccessOpen(false);
    }

    const releaseAccess = () => {
        if(detail && detail.token) {
            let cfg = new Configuration({
                headers: {
                    "X-TOKEN": user.token
                }
            });
            let shareApi = new ShareApi(cfg);
            shareApi.unaccess({body: {frontendToken: detail.token, envZId: detail.zId, shrToken: detail.shrToken}})
                .then(d => {
                    setReleaseAccessOpen(false);
                })
                .catch(e => {
                    console.log("releaseAccess", e);
                });
        }
    }

    useEffect(() => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let metadataApi = new MetadataApi(cfg);
        metadataApi.getFrontendDetail({feId: access.data.feId as number})
            .then(d => {
                delete d.id;
                setDetail(d);
            })
            .catch(e => {
                console.log("AccessPanel", e);
            })
    }, [access]);

    const customProperties = {
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString()
    }

    const labels = {
        createdAt: "Created",
        shrToken: "Share Token",
        token: "Frontend Token",
        updatedAt: "Updated",
    }

    return (
        <>
            <Typography component="div">
                <Grid2 container spacing={2}>
                    <Grid2 >
                        <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                            <Grid2 display="flex"><AccessIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                            <Grid2 display="flex" component="h3">{String(access.data.label)}</Grid2>
                        </Grid2>
                        <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                            <Tooltip title="Release Access">
                                <Button variant="contained" color="error" onClick={openReleaseAccess}><DeleteIcon /></Button>
                            </Tooltip>
                        </Grid2>
                        <Grid2 container sx={{ flexGrow: 1 }}>
                            <Grid2 display="flex">
                                <PropertyTable object={detail} custom={customProperties} labels={labels} />
                            </Grid2>
                        </Grid2>
                    </Grid2>
                </Grid2>
            </Typography>
            <ReleaseAccessModal close={closeReleaseAccess} isOpen={releaseAccessOpen} detail={detail} action={releaseAccess} />
        </>
    );
}

export default AccessPanel;
