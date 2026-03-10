import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import AccessIcon from "@mui/icons-material/Lan";
import useApiConsoleStore from "./model/store.ts";
import React, {useEffect, useState} from "react";
import {Frontend} from "./api";
import DeleteIcon from "@mui/icons-material/Delete";
import PropertyTable from "./PropertyTable.tsx";
import ReleaseAccessModal from "./ReleaseAccessModal.tsx";
import {getMetadataApi} from "./model/api.ts";
import ClipboardText from "./ClipboardText.tsx";
import BandwidthLimitedWarning from "./BandwidthLimitedWarning.tsx";
import {extractErrorMessage, isAbortError} from "./model/errors.ts";
import {PropertyRow} from "./model/util.ts";

interface AccessPanelProps {
    access: Node;
}

const AccessPanel = ({ access }: AccessPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const limited = useApiConsoleStore((state) => state.limited);
    const [detail, setDetail] = useState<Frontend>(null);
    const [errorMessage, setErrorMessage] = useState<string>("");
    const [releaseAccessOpen, setReleaseAccessOpen] = useState<boolean>(false);
    const openReleaseAccess = () => {
        setReleaseAccessOpen(true);
    }
    const closeReleaseAccess = () => {
        setReleaseAccessOpen(false);
    }

    useEffect(() => {
        const controller = new AbortController();
        getMetadataApi(user).getFrontendDetail({frontendId: access.data.feId as number}, { signal: controller.signal })
            .then(d => {
                const { id, zId, description, ...rest } = d;
                setDetail(rest as Frontend);
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "failed to load access details");
                setErrorMessage(msg);
            })
        return () => controller.abort();
    }, [access.data.feId]);

    const customProperties = {
        bindAddress: (row: PropertyRow) => <>
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{row.value as string}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ClipboardText text={row.value as string} />
                </Grid2>
            </Grid2>
        </>,
        frontendToken: (row: PropertyRow) => <>
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{row.value as string}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ClipboardText text={row.value as string} />
                </Grid2>
            </Grid2>
        </>,
        shareToken: (row: PropertyRow) => <>
            <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                <Grid2 display="flex" justifyContent="left">
                    <span>{row.value as string}</span>
                </Grid2>
                <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                    <ClipboardText text={row.value as string} />
                </Grid2>
            </Grid2>
        </>,
        createdAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString(),
        updatedAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString(),
    }

    const labels = {
        createdAt: "Created",
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
                            <h5 style={{ margin: 0 }}>A private access frontend {detail && detail.bindAddress ? <span>at <code>{detail.bindAddress}</code></span> : <span>with frontend token <code>{detail?.frontendToken}</code></span>}</h5>
                        </Grid2>
                        { errorMessage && <Typography color="error" sx={{ mb: 2 }}>{errorMessage}</Typography> }
                        { limited ? <BandwidthLimitedWarning /> : null }
                        <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                            <Tooltip title="Release Access">
                                <Button variant="contained" color="error" aria-label="Release Access" onClick={openReleaseAccess}><DeleteIcon /></Button>
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
            <ReleaseAccessModal close={closeReleaseAccess} isOpen={releaseAccessOpen} user={user} access={access} detail={detail} />
        </>
    );
}

export default AccessPanel;
