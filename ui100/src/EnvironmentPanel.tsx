import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import {useEffect, useState} from "react";
import {Configuration, Environment, EnvironmentApi, MetadataApi} from "./api";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useApiConsoleStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";
import ReleaseEnvironmentModal from "./ReleaseEnvironmentModal.tsx";

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({environment}: EnvironmentPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const [detail, setDetail] = useState<Environment>(null);
    const [releaseEnvironmentOpen, setReleaseEnvironmentOpen] = useState(false);

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
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let metadata = new MetadataApi(cfg);
        metadata.getEnvironmentDetail({envZId: environment.data!.envZId! as string})
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
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Release Environment">
                        <Button variant="contained" color="error" onClick={openReleaseEnvironment}><DeleteIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={detail} custom={customProperties} labels={labels} />
                    </Grid2>
                </Grid2>
            </Typography>
            <ReleaseEnvironmentModal close={closeReleaseEnvironment} isOpen={releaseEnvironmentOpen} user={user} environment={environment} detail={detail} />
        </>
    );
}

export default EnvironmentPanel;