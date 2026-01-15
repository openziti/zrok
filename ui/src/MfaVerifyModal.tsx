import { useState } from "react";
import { modalStyle } from "./styling/theme.ts";
import { Box, Button, Grid2, Modal, TextField, Typography, Alert, Link } from "@mui/material";
import { AccountApi } from "./api";

interface MfaVerifyModalProps {
    isOpen: boolean;
    pendingToken: string;
    onSuccess: (token: string) => void;
    onCancel: () => void;
}

const MfaVerifyModal = ({ isOpen, pendingToken, onSuccess, onCancel }: MfaVerifyModalProps) => {
    const [code, setCode] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [useRecoveryCode, setUseRecoveryCode] = useState(false);

    const handleVerify = async () => {
        if (code.length < 6) {
            setError("Please enter a valid code");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            const token = await new AccountApi().mfaAuthenticate({
                body: {
                    pendingToken: pendingToken,
                    code: code,
                }
            });
            onSuccess(token);
        } catch (e) {
            console.error("MFA authentication error:", e);
            setError("Invalid code. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === "Enter" && code.length >= 6) {
            handleVerify();
        }
    };

    const toggleRecoveryMode = () => {
        setUseRecoveryCode(!useRecoveryCode);
        setCode("");
        setError(null);
    };

    return (
        <Modal open={isOpen} onClose={onCancel}>
            <Box sx={{ ...modalStyle, width: 400 }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5">
                        <strong>Two-Factor Authentication</strong>
                    </Typography>
                </Grid2>

                <Typography variant="body1" sx={{ mb: 2 }}>
                    {useRecoveryCode
                        ? "Enter one of your recovery codes:"
                        : "Enter the 6-digit code from your authenticator app:"}
                </Typography>

                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}

                <TextField
                    fullWidth
                    label={useRecoveryCode ? "Recovery Code" : "Verification Code"}
                    value={code}
                    onChange={(e) => {
                        if (useRecoveryCode) {
                            // Recovery codes are alphanumeric with optional dash
                            setCode(e.target.value.toUpperCase().slice(0, 9));
                        } else {
                            // TOTP codes are 6 digits
                            setCode(e.target.value.replace(/\D/g, "").slice(0, 6));
                        }
                    }}
                    onKeyPress={handleKeyPress}
                    placeholder={useRecoveryCode ? "XXXX-XXXX" : "000000"}
                    inputProps={{
                        maxLength: useRecoveryCode ? 9 : 6,
                        style: { letterSpacing: "0.3em", textAlign: "center" }
                    }}
                    sx={{ mb: 2 }}
                    autoFocus
                />

                <Box sx={{ display: "flex", gap: 1, mb: 2 }}>
                    <Button
                        variant="contained"
                        onClick={handleVerify}
                        disabled={loading || code.length < 6}
                        fullWidth
                    >
                        {loading ? "Verifying..." : "Verify"}
                    </Button>
                    <Button variant="outlined" onClick={onCancel}>
                        Cancel
                    </Button>
                </Box>

                <Box sx={{ textAlign: "center" }}>
                    <Link
                        component="button"
                        variant="body2"
                        onClick={toggleRecoveryMode}
                        sx={{ cursor: "pointer" }}
                    >
                        {useRecoveryCode
                            ? "Use authenticator app instead"
                            : "Use a recovery code instead"}
                    </Link>
                </Box>
            </Box>
        </Modal>
    );
};

export default MfaVerifyModal;
