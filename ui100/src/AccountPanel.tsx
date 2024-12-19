import {Node} from "@xyflow/react";
import {Grid2, Paper, TableBody, TableContainer, Typography} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import {User} from "./model/user.ts";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";

interface AccountPanelProps {
    account: Node;
    user: User;
}

const AccountPanel = ({ account, user }: AccountPanelProps) => {
    const customProps = {
        token: row => <SecretToggle secret={row.value} />
    }

    return (
        <Typography component="div">
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccountIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex" component="h3">{String(account.data.label)}</Grid2>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex">
                    <PropertyTable object={user} custom={customProps} />
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default AccountPanel;
