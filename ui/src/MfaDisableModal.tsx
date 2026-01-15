import { User } from "./model/user.ts";
import { useEffect, useState } from "react";
import { modalStyle } from "./styling/theme.ts";
import { Box, Button, Grid2, Modal, TextField, Typography, Alert } from "@mui/material";
import { getAccountApi } from "./model/api.ts";

interface MfaDisableModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    onMfaDisabled: () => void;
}

const MfaDisableModal = ({ close, isOpen, user, onMfaDisabled }: MfaDisableModalProps) => {
    const [password, setPassword] = useState("");
    const [code, setCode] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    useEffect(() => {
        if (isOpen) {
            setPassword("");
            setCode("");
            setError(null);
            setSuccess(false);
        }
    }, [isOpen]);

    const handleDisable = async () => {
        if (!password || code.length < 6) {
            setError("Please enter your password and verification code");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            await getAccountApi(user).mfaDisable({
                body: {
                    password: password,
                    code: code,
                }
            });
            setSuccess(true);
            onMfaDisabled();
            setTimeout(() => close(), 2000);
        } catch (e) {
            console.error("MFA disable error:", e);
            setError("Failed to disable MFA. Check your password and code.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle, width: 450 }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5">
                        <strong>Disable Two-Factor Authentication</strong>
                    </Typography>
                </Grid2>

                {success ? (
                    <Alert severity="success" sx={{ mt: 2 }}>
                        Two-factor authentication has been disabled.
                    </Alert>
                ) : (
                    <>
                        <Alert severity="warning" sx={{ mt: 2, mb: 2 }}>
                            Disabling MFA will make your account less secure. Are you sure you want to continue?
                        </Alert>

                        {error && (
                            <Alert severity="error" sx={{ mb: 2 }}>
                                {error}
                            </Alert>
                        )}

                        <TextField
                            fullWidth
                            type="password"
                            label="Current Password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            sx={{ mb: 2 }}
                        />

                        <TextField
                            fullWidth
                            label="Verification Code"
                            value={code}
                            onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                            placeholder="000000"
                            inputProps={{
                                maxLength: 6,
                                style: { letterSpacing: "0.5em", textAlign: "center" }
                            }}
                            helperText="Enter the 6-digit code from your authenticator app"
                            sx={{ mb: 2 }}
                        />

                        <Box sx={{ display: "flex", gap: 1 }}>
                            <Button
                                variant="contained"
                                color="error"
                                onClick={handleDisable}
                                disabled={loading || !password || code.length < 6}
                            >
                                {loading ? "Disabling..." : "Disable MFA"}
                            </Button>
                            <Button variant="outlined" onClick={close}>
                                Cancel
                            </Button>
                        </Box>
                    </>
                )}
            </Box>
        </Modal>
    );
};

export default MfaDisableModal;
