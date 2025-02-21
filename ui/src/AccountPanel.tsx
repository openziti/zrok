import {Node} from "@xyflow/react";
import {Button, Grid2, Tooltip, Typography} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useApiConsoleStore from "./model/store.ts";
import PasswordIcon from "@mui/icons-material/Password";
import TokenIcon from "@mui/icons-material/Key";
import React, {useState} from "react";
import AccountPasswordChangeModal from "./AccountPasswordChangeModal.tsx";
import RegenerateAccountTokenModal from "./RegenerateAccountTokenModal.tsx";
import ClipboardText from "./ClipboardText.tsx";
import AccountMetricsModal from "./AccountMetricsModal.tsx";
import MetricsIcon from "@mui/icons-material/QueryStats";
import BandwidthLimitedWarning from "./BandwidthLimitedWarning.tsx";

interface AccountPanelProps {
    account: Node;
}

const AccountPanel = ({ account }: AccountPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const limited = useApiConsoleStore((state) => state.limited);
    const [accountMetricsOpen, setAccountMetricsOpen] = useState<boolean>(false);
    const openAccountMetrics = () => {
        setAccountMetricsOpen(true);
    }
    const closeAccountMetrics = () => {
        setAccountMetricsOpen(false);
    }
    const [changePasswordOpen, setChangePasswordOpen] = useState<boolean>(false);
    const openChangePassword = () => {
        setChangePasswordOpen(true);
    }
    const closeChangePassword = () => {
        setChangePasswordOpen(false);
    }
    const [regenerateOpen, setRegenerateOpen] = useState<boolean>(false);
    const openRegenerate = () => {
        setRegenerateOpen(true);
    }
    const closeRegenerate = () => {
        setRegenerateOpen(false);
    }

    const customProperties = {
        token: row => <Grid2 container><SecretToggle secret={row.value} /><ClipboardText text={row.value} /></Grid2>
    }

    const label = {
        token: "Account Token"
    }

    return (
        <>
            <Typography component="div">
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Grid2 display="flex"><AccountIcon sx={{ fontSize: 30, mr: 0.5 }}/></Grid2>
                    <Grid2 display="flex" component="h3">{String(account.data.label)}</Grid2>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, mt: 0, mb: 2 }} alignItems="center">
                    <h5 style={{ margin: 0 }}>Your zrok account <code>{user.email}</code></h5>
                </Grid2>
                { limited ? <BandwidthLimitedWarning /> : null }
                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Account Metrics">
                        <Button variant="contained" onClick={openAccountMetrics}><MetricsIcon /></Button>
                    </Tooltip>
                    <Tooltip title="Change Password" sx={{ ml: 1 }}>
                        <Button variant="contained" color="error" onClick={openChangePassword}><PasswordIcon /></Button>
                    </Tooltip>
                    <Tooltip title="Regenerate Account Token" sx={{ ml: 1 }}>
                        <Button variant="contained" color="error" onClick={openRegenerate}><TokenIcon /></Button>
                    </Tooltip>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={user} custom={customProperties} labels={label} />
                    </Grid2>
                </Grid2>
            </Typography>
            <AccountMetricsModal close={closeAccountMetrics} isOpen={accountMetricsOpen} user={user} />
            <AccountPasswordChangeModal close={closeChangePassword} isOpen={changePasswordOpen} user={user} />
            <RegenerateAccountTokenModal close={closeRegenerate} isOpen={regenerateOpen} user={user} />
        </>
    );
}

export default AccountPanel;
