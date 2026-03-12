import {Node} from "@xyflow/react";
import {Button, CircularProgress, Grid2, Tooltip, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import {Share} from "./api";
import React, {useEffect, useState} from "react";
import PropertyTable from "./PropertyTable.tsx";
import useApiConsoleStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import ReleaseShareModal from "./ReleaseShareModal.tsx";
import {getMetadataApi} from "./model/api.ts";
import ClipboardText from "./ClipboardText.tsx";
import MetricsIcon from "@mui/icons-material/QueryStats";
import ShareMetricsModal from "./ShareMetricsModal.tsx";
import BandwidthLimitedWarning from "./BandwidthLimitedWarning.tsx";
import {extractErrorMessage, isAbortError} from "./model/errors.ts";
import {PropertyRow} from "./model/util.ts";

interface SharePanelProps {
    share: Node;
}

const SharePanel = ({ share }: SharePanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const [detail, setDetail] = useState<Share>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [errorMessage, setErrorMessage] = useState<string>("");
    const [shareMetricsOpen, setShareMetricsOpen] = useState<boolean>(false);
    const openShareMetrics = () => {
        setShareMetricsOpen(true);
    }
    const closeShareMetrics = () => {
        setShareMetricsOpen(false);
    }
    const [releaseShareOpen, setReleaseShareOpen] = useState<boolean>(false);
    const openReleaseShare = () => {
        setReleaseShareOpen(true);
    }
    const closeReleaseShare = () => {
        setReleaseShareOpen(false);
    }

    const customProperties = {
        createdAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString(),
        updatedAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString(),
        reserved: (row: PropertyRow) => row.value ? "reserved" : "ephemeral",
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
        frontendEndpoints: (row: PropertyRow) => {
            if (!row.value || row.value.length === 0) {
                return <span>None</span>;
            }
            return (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                    {row.value.map((endpoint, index) => (
                        <Grid2 key={index} container sx={{ flexGrow: 1 }} alignItems="center">
                            <Grid2 display="flex" justifyContent="left">
                                <span>{endpoint}</span>
                            </Grid2>
                            <Grid2 display="flex" justifyContent="right" sx={{ flexGrow: 1 }}>
                                <ClipboardText text={endpoint} />
                            </Grid2>
                        </Grid2>
                    ))}
                </div>
            );
        }
    }

    const labels = {
        createdAt: "Created",
        token: "Share Token",
        updatedAt: "Updated"
    }

    useEffect(() => {
        if (!user) return;
        setLoading(true);
        const controller = new AbortController();
        getMetadataApi(user).getShareDetail({ shareToken: share.data!.shareToken! as string }, { signal: controller.signal })
            .then(d => {
                const { activity, limited, zId, ...rest } = d;
                if(rest.shareMode === "private") {
                    const { frontendEndpoints, ...withoutEndpoints } = rest;
                    setDetail(withoutEndpoints as Share);
                } else {
                    setDetail(rest as Share);
                }
                setLoading(false);
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "failed to load share details");
                setErrorMessage(msg);
                setLoading(false);
            })
        return () => controller.abort();
    }, [share.id]);

    if (!user) return null;

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 0 }} alignItems="center">
                    <Grid2 display="flex"><ShareIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h4">{String(share.data.label)}</Grid2>
                </Grid2>
                {loading ? (
                    <Grid2 container justifyContent="center" sx={{ mt: 4 }}><CircularProgress /></Grid2>
                ) : (
                    <>
                        <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2 }} alignItems="center">
                            <h5 style={{ margin: 0 }}>A {detail ? detail.shareMode : ''}{detail && detail.reserved ? ', reserved ' : ''} {detail?.backendMode} share with the share token <code>{share.id}</code></h5>
                        </Grid2>
                        { errorMessage && <Typography color="error" sx={{ mb: 2 }}>{errorMessage}</Typography> }
                        { share.data.limited ? <BandwidthLimitedWarning /> : null }
                        <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                            <Tooltip title="Share Metrics">
                                <Button variant="contained" aria-label="Share Metrics" onClick={openShareMetrics}><MetricsIcon /></Button>
                            </Tooltip>
                            <Tooltip title="Release Share" sx={{ ml: 1 }}>
                                <Button variant="contained" color="error" aria-label="Release Share" onClick={openReleaseShare}><DeleteIcon /></Button>
                            </Tooltip>
                        </Grid2>
                        <Grid2 container sx={{ flexGrow: 1 }}>
                            <Grid2 display="flex">
                                <PropertyTable object={detail} custom={customProperties} labels={labels} />
                            </Grid2>
                        </Grid2>
                    </>
                )}
            </Typography>
            <ShareMetricsModal close={closeShareMetrics} isOpen={shareMetricsOpen} user={user} share={share} />
            <ReleaseShareModal close={closeReleaseShare} isOpen={releaseShareOpen} user={user} share={share} detail={detail} />
        </>
    );
}

export default SharePanel;
