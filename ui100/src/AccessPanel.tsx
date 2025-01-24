import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import AccessIcon from "@mui/icons-material/Lan";
import useStore from "./model/store.ts";
import {useEffect, useState} from "react";
import {Configuration, Frontend, MetadataApi, ShareApi} from "./api";
import DeleteIcon from "@mui/icons-material/Delete";
import PropertyTable from "./PropertyTable.tsx";
import ReleaseAccessModal from "./ReleaseAccessModal.tsx";

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
            shareApi.unaccess({body: {frontendToken: detail.token, envZId: access.data.envZId as string, shrToken: detail.shrToken}})
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
                delete d.zId;
                setDetail(d);
            })
            .catch(e => {
                console.log("AccessPanel", e);
            })
    }, [access]);

    const customProperties = {
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString(),
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
                        <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                            <Grid2 display="flex"><AccessIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                            <Grid2 display="flex" component="h3">{String(access.data.label)}</Grid2>
                        </Grid2>
                        <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2, p: 0 }} alignItems="center">
                            <h5 style={{ margin: 0 }}>A private access frontend with the token <code>{access.id}</code></h5>
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
