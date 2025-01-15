import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useStore from "./model/store.ts";
import PasswordIcon from "@mui/icons-material/Password";
import TokenIcon from "@mui/icons-material/Key";

interface AccountPanelProps {
    account: Node;
}

const AccountPanel = ({ account}: AccountPanelProps) => {
    const user = useStore((state) => state.user);

    const customProps = {
        token: row => <SecretToggle secret={row.value} />
    }

    const label = {
        token: "Account Token"
    }

    return (
        <Typography component="div">
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccountIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex" component="h3">{String(account.data.label)}</Grid2>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                <Tooltip title="Change Password">
                    <Button variant="contained" color="error"><PasswordIcon /></Button>
                </Tooltip>
                <Tooltip title="Regenerate Account Token" sx={{ ml: 1 }}>
                    <Button variant="contained" color="error"><TokenIcon /></Button>
                </Tooltip>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1 }}>
                <Grid2 display="flex">
                    <PropertyTable object={user} custom={customProps} labels={label} />
                </Grid2>
            </Grid2>
        </Typography>
    );
}

export default AccountPanel;
