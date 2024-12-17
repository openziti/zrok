import {Node} from "@xyflow/react";
import {Card, Grid2, Typography} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";

interface SharePanelProps {
    share: Node;
}

const SharePanel = ({ share }: SharePanelProps) => {
    return (
        <Typography component="div">
            <Grid2 container spacing={2}>
                <Grid2 >
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Grid2 display="flex"><ShareIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                        <Grid2 display="flex" component="h3">{String(share.data.label)}</Grid2>
                    </Grid2>
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default SharePanel;
