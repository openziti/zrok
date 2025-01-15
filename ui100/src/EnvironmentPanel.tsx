import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import {useEffect, useState} from "react";
import {Configuration, Environment, MetadataApi} from "./api";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useStore from "./model/store.ts";
import DeleteIcon from "@mui/icons-material/Delete";

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({ environment }: EnvironmentPanelProps) => {
    const user = useStore((state) => state.user);
    const [detail, setDetail] = useState<Environment>(null);

    const customProperties = {
        zId: row => <SecretToggle secret={row.value} />,
        createdAt: row => new Date(row.value).toLocaleString(),
        updatedAt: row => new Date(row.value).toLocaleString()
    }

    const labels = {
        zId: "OpenZiti Service"
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
                setDetail(env);
            })
            .catch(e => {
                console.log("EnvironmentPanel", e);
            })
    }, [environment]);

    return (
        <Typography component="div">
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex" component="h3">{String(environment.data.label)}</Grid2>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                <Tooltip title="Release Environment">
                    <Button variant="contained" color="error"><DeleteIcon /></Button>
                </Tooltip>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex">
                    <PropertyTable object={detail} custom={customProperties} labels={labels} />
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default EnvironmentPanel;