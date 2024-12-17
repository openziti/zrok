import {Node} from "@xyflow/react";
import {Grid2, Typography} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";

interface EnvironmentPanelProps {
    environment: Node;
}

const EnvironmentPanel = ({ environment }: EnvironmentPanelProps) => {
    return (
        <Typography component="div">
            <Grid2 container spacing={2}>
                <Grid2 >
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                        <Grid2 display="flex" component="h3">{String(environment.data.label)}</Grid2>
                    </Grid2>
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default EnvironmentPanel;