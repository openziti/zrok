import {Node} from "@xyflow/react";
import {Button, Chip, Grid2, Tooltip, Typography} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import PropertyTable from "./PropertyTable.tsx";
import SecretToggle from "./SecretToggle.tsx";
import useApiConsoleStore from "./model/store.ts";
import PasswordIcon from "@mui/icons-material/Password";
import TokenIcon from "@mui/icons-material/Key";
import SecurityIcon from "@mui/icons-material/Security";
import KeyOffIcon from "@mui/icons-material/KeyOff";
import VpnKeyIcon from "@mui/icons-material/VpnKey";
import React, {useEffect, useState} from "react";
import AccountPasswordChangeModal from "./AccountPasswordChangeModal.tsx";
import RegenerateAccountTokenModal from "./RegenerateAccountTokenModal.tsx";
import ClipboardText from "./ClipboardText.tsx";
import AccountMetricsModal from "./AccountMetricsModal.tsx";
import MetricsIcon from "@mui/icons-material/QueryStats";
import BandwidthLimitedWarning from "./BandwidthLimitedWarning.tsx";
import MfaSetupModal from "./MfaSetupModal.tsx";
import MfaDisableModal from "./MfaDisableModal.tsx";
import MfaRecoveryCodesModal from "./MfaRecoveryCodesModal.tsx";
import {getAccountApi} from "./model/api.ts";

interface AccountPanelProps {
    account: Node;
}

const AccountPanel = ({ account }: AccountPanelProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const limited = useApiConsoleStore((state) => state.limited);

    // MFA state
    const [mfaEnabled, setMfaEnabled] = useState<boolean>(false);
    const [recoveryCodesRemaining, setRecoveryCodesRemaining] = useState<number>(0);
    const [mfaLoading, setMfaLoading] = useState<boolean>(true);

    // Modal states
    const [accountMetricsOpen, setAccountMetricsOpen] = useState<boolean>(false);
    const [changePasswordOpen, setChangePasswordOpen] = useState<boolean>(false);
    const [regenerateOpen, setRegenerateOpen] = useState<boolean>(false);
    const [mfaSetupOpen, setMfaSetupOpen] = useState<boolean>(false);
    const [mfaDisableOpen, setMfaDisableOpen] = useState<boolean>(false);
    const [mfaRecoveryCodesOpen, setMfaRecoveryCodesOpen] = useState<boolean>(false);

    // Fetch MFA status on mount
    useEffect(() => {
        fetchMfaStatus();
    }, []);

    const fetchMfaStatus = async () => {
        setMfaLoading(true);
        try {
            const status = await getAccountApi(user).mfaStatus();
            setMfaEnabled(status.enabled || false);
            setRecoveryCodesRemaining(status.recoveryCodesRemaining || 0);
        } catch (e) {
            console.error("Failed to fetch MFA status:", e);
            setMfaEnabled(false);
        } finally {
            setMfaLoading(false);
        }
    };

    const handleMfaEnabled = () => {
        fetchMfaStatus();
    };

    const handleMfaDisabled = () => {
        setMfaEnabled(false);
        setRecoveryCodesRemaining(0);
    };

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

                {/* MFA Status */}
                <Grid2 container sx={{ flexGrow: 1, mb: 2 }} alignItems="center">
                    <Typography variant="body2" sx={{ mr: 1 }}>Two-Factor Authentication:</Typography>
                    {mfaLoading ? (
                        <Chip label="Loading..." size="small" />
                    ) : mfaEnabled ? (
                        <Chip label="Enabled" color="success" size="small" icon={<SecurityIcon />} />
                    ) : (
                        <Chip label="Disabled" color="default" size="small" />
                    )}
                    {mfaEnabled && recoveryCodesRemaining > 0 && (
                        <Typography variant="caption" sx={{ ml: 2, color: recoveryCodesRemaining <= 3 ? "warning.main" : "text.secondary" }}>
                            {recoveryCodesRemaining} recovery codes remaining
                        </Typography>
                    )}
                </Grid2>

                <Grid2 container sx={{ flexGrow: 1, mb: 3 }} alignItems="left">
                    <Tooltip title="Account Metrics">
                        <Button variant="contained" onClick={() => setAccountMetricsOpen(true)}><MetricsIcon /></Button>
                    </Tooltip>
                    <Tooltip title="Change Password" sx={{ ml: 1 }}>
                        <Button variant="contained" color="error" onClick={() => setChangePasswordOpen(true)}><PasswordIcon /></Button>
                    </Tooltip>
                    <Tooltip title="Regenerate Account Token" sx={{ ml: 1 }}>
                        <Button variant="contained" color="error" onClick={() => setRegenerateOpen(true)}><TokenIcon /></Button>
                    </Tooltip>

                    {/* MFA Buttons */}
                    {!mfaLoading && !mfaEnabled && (
                        <Tooltip title="Enable Two-Factor Authentication" sx={{ ml: 1 }}>
                            <Button variant="contained" color="success" onClick={() => setMfaSetupOpen(true)}>
                                <SecurityIcon />
                            </Button>
                        </Tooltip>
                    )}
                    {!mfaLoading && mfaEnabled && (
                        <>
                            <Tooltip title="View Recovery Codes" sx={{ ml: 1 }}>
                                <Button variant="contained" color="warning" onClick={() => setMfaRecoveryCodesOpen(true)}>
                                    <VpnKeyIcon />
                                </Button>
                            </Tooltip>
                            <Tooltip title="Disable Two-Factor Authentication" sx={{ ml: 1 }}>
                                <Button variant="contained" color="error" onClick={() => setMfaDisableOpen(true)}>
                                    <KeyOffIcon />
                                </Button>
                            </Tooltip>
                        </>
                    )}
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex">
                        <PropertyTable object={user} custom={customProperties} labels={label} />
                    </Grid2>
                </Grid2>
            </Typography>
            <AccountMetricsModal close={() => setAccountMetricsOpen(false)} isOpen={accountMetricsOpen} user={user} />
            <AccountPasswordChangeModal close={() => setChangePasswordOpen(false)} isOpen={changePasswordOpen} user={user} />
            <RegenerateAccountTokenModal close={() => setRegenerateOpen(false)} isOpen={regenerateOpen} user={user} />
            <MfaSetupModal close={() => setMfaSetupOpen(false)} isOpen={mfaSetupOpen} user={user} onMfaEnabled={handleMfaEnabled} />
            <MfaDisableModal close={() => setMfaDisableOpen(false)} isOpen={mfaDisableOpen} user={user} onMfaDisabled={handleMfaDisabled} />
            <MfaRecoveryCodesModal close={() => setMfaRecoveryCodesOpen(false)} isOpen={mfaRecoveryCodesOpen} user={user} recoveryCodesRemaining={recoveryCodesRemaining} />
        </>
    );
}

export default AccountPanel;
