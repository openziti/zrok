import {Node} from "@xyflow/react";
import {Grid2, Typography} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";

interface AccountPanelProps {
    account: Node;
}

const AccountPanel = ({ account }: AccountPanelProps) => {
    return (
        <Typography component="div">
            <Grid2 container spacing={2}>
                <Grid2 >
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Grid2 display="flex"><AccountIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                        <Grid2 display="flex" component="h3">{String(account.data.label)}</Grid2>
                    </Grid2>
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default AccountPanel;
