import {User} from "./model/user.ts";
import {useEffect, useState} from "react";
import {buildMetrics} from "./model/util.ts";
import {getMetadataApi} from "./model/api.ts";
import {Node} from "@xyflow/react";
import {Box, Grid2, Modal, Typography} from "@mui/material";
import {modalStyle} from "./styling/theme.ts";
import MetricsGraph from "./MetricsGraph.tsx";
import {extractErrorMessage, isAbortError} from "./model/errors.ts";

interface EnvironmentMetricsModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    environment: Node;
}

const EnvironmentMetricsModal = ({ close, isOpen, user, environment }: EnvironmentMetricsModalProps) => {
    const [metrics30, setMetrics30] = useState(buildMetrics({}));
    const [metrics7, setMetrics7] = useState(buildMetrics({}));
    const [metrics1, setMetrics1] = useState(buildMetrics({}));
    const [errorMessage, setErrorMessage] = useState<string>("");

    useEffect(() => {
        if (!isOpen) return;
        const controller = new AbortController();
        setErrorMessage("");
        let metadataApi = getMetadataApi(user);
        metadataApi.getEnvironmentMetrics({envId: String(environment.data.envZId) }, { signal: controller.signal })
            .then(d => {
                setMetrics30(buildMetrics(d));
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "unable to load metrics");
                setErrorMessage(msg);
            });
        metadataApi.getAccountMetrics({envId: String(environment.data.envZId), duration: "168h"}, { signal: controller.signal })
            .then(d => {
                setMetrics7(buildMetrics(d));
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "unable to load metrics");
                setErrorMessage(msg);
            });
        metadataApi.getAccountMetrics({envId: String(environment.data.envZId), duration: "24h"}, { signal: controller.signal })
            .then(d => {
                setMetrics1(buildMetrics(d));
            })
            .catch(async (e) => {
                if (isAbortError(e)) return;
                const msg = await extractErrorMessage(e, "unable to load metrics");
                setErrorMessage(msg);
            });
        return () => controller.abort();
    }, [isOpen, environment]);

    return (
        <Modal open={isOpen} onClose={close} aria-labelledby="modal-title-environment-metrics">
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5" id="modal-title-environment-metrics"><strong>Environment Metrics</strong></Typography>
                </Grid2>
                { errorMessage && <Grid2 container sx={{ flexGrow: 1, p: 1 }}><Typography color="error">{errorMessage}</Typography></Grid2> }
                <MetricsGraph title="Last 30 Days" data={metrics30.data} />
                <MetricsGraph title="Last 7 Days" data={metrics7.data} showTime />
                <MetricsGraph title="Last 24 Hours" data={metrics1.data} showTime />
            </Box>
        </Modal>
    );
}

export default EnvironmentMetricsModal;