import {Node} from "@xyflow/react";
import {Grid2, Typography} from "@mui/material";
import AccessIcon from "@mui/icons-material/Lan";

interface AccessPanelProps {
    access: Node;
}

const AccessPanel = ({ access }: AccessPanelProps) => {
    return (
        <Typography component="div">
            <Grid2 container spacing={2}>
                <Grid2 >
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Grid2 display="flex"><AccessIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                        <Grid2 display="flex" component="h3">{String(access.data.label)}</Grid2>
                    </Grid2>
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default AccessPanel;
