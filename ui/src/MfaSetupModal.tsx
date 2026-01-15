import { User } from "./model/user.ts";
import { useEffect, useState } from "react";
import { modalStyle } from "./styling/theme.ts";
import { Box, Button, Grid2, Modal, TextField, Typography, Stepper, Step, StepLabel, Alert } from "@mui/material";
import { getAccountApi } from "./model/api.ts";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";

interface MfaSetupModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    onMfaEnabled: () => void;
}

const MfaSetupModal = ({ close, isOpen, user, onMfaEnabled }: MfaSetupModalProps) => {
    const [activeStep, setActiveStep] = useState(0);
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);

    // Step 1: QR code data
    const [qrCode, setQrCode] = useState<string>("");
    const [secret, setSecret] = useState<string>("");

    // Step 2: Verification code
    const [verificationCode, setVerificationCode] = useState("");

    // Step 3: Recovery codes
    const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
    const [codesAcknowledged, setCodesAcknowledged] = useState(false);

    const steps = ["Scan QR Code", "Verify Code", "Save Recovery Codes"];

    useEffect(() => {
        if (isOpen) {
            setActiveStep(0);
            setError(null);
            setQrCode("");
            setSecret("");
            setVerificationCode("");
            setRecoveryCodes([]);
            setCodesAcknowledged(false);
            initiateSetup();
        }
    }, [isOpen]);

    const initiateSetup = async () => {
        setLoading(true);
        setError(null);
        try {
            const result = await getAccountApi(user).mfaSetup();
            setQrCode(result.qrCode || "");
            setSecret(result.secret || "");
        } catch (e) {
            console.error("MFA setup error:", e);
            setError("Failed to initiate MFA setup. MFA may already be enabled.");
        } finally {
            setLoading(false);
        }
    };

    const verifyCode = async () => {
        if (verificationCode.length < 6) {
            setError("Please enter a valid 6-digit code");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            const result = await getAccountApi(user).mfaVerify({ body: { code: verificationCode } });
            setRecoveryCodes(result.recoveryCodes || []);
            setActiveStep(2);
        } catch (e) {
            console.error("MFA verification error:", e);
            setError("Invalid verification code. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    const copySecret = () => {
        navigator.clipboard.writeText(secret);
    };

    const copyRecoveryCodes = () => {
        const codesText = recoveryCodes.join("\n");
        navigator.clipboard.writeText(codesText);
    };

    const downloadRecoveryCodes = () => {
        const codesText = `zrok Recovery Codes for ${user.email}\n` +
            `Generated: ${new Date().toISOString()}\n\n` +
            `Keep these codes safe. Each code can only be used once.\n\n` +
            recoveryCodes.join("\n");

        const blob = new Blob([codesText], { type: "text/plain" });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = "zrok-recovery-codes.txt";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const handleClose = () => {
        if (activeStep === 2 && recoveryCodes.length > 0) {
            onMfaEnabled();
        }
        close();
    };

    const renderStep0 = () => (
        <Box>
            <Typography variant="body1" sx={{ mb: 2 }}>
                Scan this QR code with your authenticator app (Google Authenticator, Authy, etc.)
            </Typography>
            {qrCode && (
                <Box sx={{ textAlign: "center", mb: 2 }}>
                    <img src={qrCode} alt="MFA QR Code" style={{ maxWidth: "200px" }} />
                </Box>
            )}
            <Typography variant="body2" sx={{ mb: 1 }}>
                Can't scan? Enter this code manually:
            </Typography>
            <Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
                <TextField
                    fullWidth
                    value={secret}
                    InputProps={{ readOnly: true }}
                    size="small"
                />
                <Button onClick={copySecret} sx={{ ml: 1 }}>
                    <ContentCopyIcon />
                </Button>
            </Box>
            <Button
                variant="contained"
                onClick={() => setActiveStep(1)}
                disabled={!secret}
            >
                Next
            </Button>
        </Box>
    );

    const renderStep1 = () => (
        <Box>
            <Typography variant="body1" sx={{ mb: 2 }}>
                Enter the 6-digit code from your authenticator app to verify setup:
            </Typography>
            <TextField
                fullWidth
                label="Verification Code"
                value={verificationCode}
                onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                placeholder="000000"
                inputProps={{ maxLength: 6, style: { letterSpacing: "0.5em", textAlign: "center" } }}
                sx={{ mb: 2 }}
            />
            <Box sx={{ display: "flex", gap: 1 }}>
                <Button variant="outlined" onClick={() => setActiveStep(0)}>
                    Back
                </Button>
                <Button
                    variant="contained"
                    onClick={verifyCode}
                    disabled={loading || verificationCode.length < 6}
                >
                    {loading ? "Verifying..." : "Verify"}
                </Button>
            </Box>
        </Box>
    );

    const renderStep2 = () => (
        <Box>
            <Alert severity="warning" sx={{ mb: 2 }}>
                Save these recovery codes in a safe place. You'll need them if you lose access to your authenticator app.
                Each code can only be used once.
            </Alert>
            <Box sx={{
                bgcolor: "#f5f5f5",
                p: 2,
                borderRadius: 1,
                fontFamily: "monospace",
                mb: 2
            }}>
                {recoveryCodes.map((code, index) => (
                    <Typography key={index} sx={{ fontFamily: "monospace" }}>
                        {code}
                    </Typography>
                ))}
            </Box>
            <Box sx={{ display: "flex", gap: 1, mb: 2 }}>
                <Button variant="outlined" onClick={copyRecoveryCodes}>
                    Copy All
                </Button>
                <Button variant="outlined" onClick={downloadRecoveryCodes}>
                    Download
                </Button>
            </Box>
            <Box sx={{ mb: 2 }}>
                <label style={{ display: "flex", alignItems: "center", cursor: "pointer" }}>
                    <input
                        type="checkbox"
                        checked={codesAcknowledged}
                        onChange={(e) => setCodesAcknowledged(e.target.checked)}
                        style={{ marginRight: "8px" }}
                    />
                    I have saved my recovery codes
                </label>
            </Box>
            <Button
                variant="contained"
                onClick={handleClose}
                disabled={!codesAcknowledged}
            >
                Done
            </Button>
        </Box>
    );

    return (
        <Modal open={isOpen} onClose={activeStep < 2 ? close : undefined}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Enable Two-Factor Authentication</strong></Typography>
                </Grid2>

                <Stepper activeStep={activeStep} sx={{ mb: 3 }}>
                    {steps.map((label) => (
                        <Step key={label}>
                            <StepLabel>{label}</StepLabel>
                        </Step>
                    ))}
                </Stepper>

                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}

                {loading && activeStep === 0 ? (
                    <Typography>Loading...</Typography>
                ) : (
                    <>
                        {activeStep === 0 && renderStep0()}
                        {activeStep === 1 && renderStep1()}
                        {activeStep === 2 && renderStep2()}
                    </>
                )}
            </Box>
        </Modal>
    );
};

export default MfaSetupModal;
