import {Node} from "@xyflow/react";
import {Box, Button, CircularProgress, Grid2, Tooltip, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import React, {useCallback, useEffect, useState} from "react";
import {Environment} from "./api";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useApiConsoleStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import ReleaseEnvironmentModal from "./ReleaseEnvironmentModal.tsx";
import {getMetadataApi} from "./model/api.ts";
import MetricsIcon from "@mui/icons-material/QueryStats";
import EnvironmentMetricsModal from "./EnvironmentMetricsModal.tsx";
import BandwidthLimitedWarning from "./BandwidthLimitedWarning.tsx";
import {extractErrorMessage, isAbortError} from "./model/errors.ts";
import {PropertyRow} from "./model/util.ts";

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({environment}: EnvironmentPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const environmentId = environment.data?.envZId as string | undefined;
    const [detail, setDetail] = useState<Environment>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [errorMessage, setErrorMessage] = useState<string>("");
    const [environmentMetricsOpen, setEnvironmentMetricsOpen] = useState<boolean>(false);
    const openEnvironmentMetrics = () => {
        setEnvironmentMetricsOpen(true);
    }
    const closeEnvironmentMetrics = () => {
        setEnvironmentMetricsOpen(false);
    }
    const [releaseEnvironmentOpen, setReleaseEnvironmentOpen] = useState<boolean>(false);
    const openReleaseEnvironment = () => {
        setReleaseEnvironmentOpen(true);
    }
    const closeReleaseEnvironment = () => {
        setReleaseEnvironmentOpen(false);
    }

    const customProperties = {
        zId: (row: PropertyRow) => <SecretToggle secret={row.value as string} />,
        createdAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString(),
        updatedAt: (row: PropertyRow) => new Date(row.value as string).toLocaleString()
    }

    const labels = {
        createdAt: "Created",
        updatedAt: "Updated"
    };

    const loadDetail = useCallback((signal?: AbortSignal) => {
        if (!user || !environmentId) {
            return Promise.resolve();
        }

        setLoading(true);
        setErrorMessage("");
        setDetail(null);
        return getMetadataApi(user).getEnvironmentDetail({envZId: environmentId}, { signal })
            .then(d => {
                const nextDetail = {...d.environment!} as Partial<Environment> & Record<string, unknown>;
                delete nextDetail.activity;
                delete nextDetail.limited;
                delete nextDetail.zId;
                setDetail(nextDetail as Environment);
                setLoading(false);
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "failed to load environment details");
                setErrorMessage(msg);
                setLoading(false);
            });
    }, [environmentId, user]);

    useEffect(() => {
        if (!user || !environmentId) return;
        const controller = new AbortController();
        void loadDetail(controller.signal);
        return () => controller.abort();
    }, [environmentId, loadDetail, user]);

    if (!user) return null;

    const actionsDisabled = detail == null;

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h3">{String(environment.data.label)}</Grid2>
                </Grid2>
                {loading ? (
                    <Grid2 container justifyContent="center" sx={{ mt: 4 }}><CircularProgress /></Grid2>
                ) : (
                    <>
                        { errorMessage && <Typography color="error" sx={{ mb: 2 }}>{errorMessage}</Typography> }
                        { environment.data.limited ? <BandwidthLimitedWarning /> : null }
                        <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                            <Tooltip title="Environment Metrics">
                                <span>
                                    <Button variant="contained" aria-label="Environment Metrics" onClick={openEnvironmentMetrics} disabled={actionsDisabled}><MetricsIcon /></Button>
                                </span>
                            </Tooltip>
                            <Box sx={{ ml: 1, display: "inline-flex" }}>
                                <Tooltip title="Release Environment">
                                    <span>
                                        <Button variant="contained" color="error" aria-label="Release Environment" onClick={openReleaseEnvironment} disabled={actionsDisabled}><DeleteIcon /></Button>
                                    </span>
                                </Tooltip>
                            </Box>
                            {actionsDisabled ? <Button variant="outlined" sx={{ ml: 1 }} onClick={() => void loadDetail()}>Retry</Button> : null}
                        </Grid2>
                        {detail ? (
                            <>
                                <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2, p: 0 }} alignItems="center">
                                    <Box component="h5" sx={{ m: 0 }}>An environment on a host with address <code>{detail.address}</code></Box>
                                </Grid2>
                                <Grid2 container sx={{ flexGrow: 1 }}>
                                    <Grid2 display="flex">
                                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                                    </Grid2>
                                </Grid2>
                            </>
                        ) : null}
                    </>
                )}
            </Typography>
            <EnvironmentMetricsModal close={closeEnvironmentMetrics} isOpen={environmentMetricsOpen} user={user} environment={environment} />
            <ReleaseEnvironmentModal close={closeReleaseEnvironment} isOpen={releaseEnvironmentOpen} user={user} environment={environment} detail={detail} />
        </>
    );
};

export default EnvironmentPanel;
