import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import React, {useEffect, useState} from "react";
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

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({environment}: EnvironmentPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const [detail, setDetail] = useState<Environment>(null);
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
        zId: row => <SecretToggle secret={row.value} />,
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString()
    }

    const labels = {
        createdAt: "Created",
        updatedAt: "Updated"
    }

    useEffect(() => {
        getMetadataApi(user).getEnvironmentDetail({envZId: environment.data!.envZId! as string})
            .then(d => {
                let env = d.environment!;
                delete env.activity;
                delete env.limited;
                delete env.zId;
                setDetail(env);
            })
            .catch(e => {
                console.log("EnvironmentPanel", e);
            })
    }, [environment]);

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h3">{String(environment.data.label)}</Grid2>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2, p: 0 }} alignItems="center">
                    <h5 style={{ margin: 0 }}>An environment on a host with address <code>{detail ? detail.address : ''}</code></h5>
                </Grid2>
                { environment.data.limited ? <BandwidthLimitedWarning /> : null }
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Environment Metrics">
                        <Button variant="contained" onClick={openEnvironmentMetrics}><MetricsIcon /></Button>
                    </Tooltip>
                    <Tooltip title="Release Environment" sx={{ ml: 1 }}>
                        <Button variant="contained" color="error" onClick={openReleaseEnvironment}><DeleteIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                    </Grid2>
                </Grid2>
            </Typography>
            <EnvironmentMetricsModal close={closeEnvironmentMetrics} isOpen={environmentMetricsOpen} user={user} environment={environment} />
            <ReleaseEnvironmentModal close={closeReleaseEnvironment} isOpen={releaseEnvironmentOpen} user={user} environment={environment} detail={detail} />
        </>
    );
}

export default EnvironmentPanel;