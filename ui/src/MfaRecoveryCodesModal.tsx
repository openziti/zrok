import { User } from "./model/user.ts";
import { useEffect, useState } from "react";
import { modalStyle } from "./styling/theme.ts";
import { Box, Button, Grid2, Modal, TextField, Typography, Alert } from "@mui/material";
import { getAccountApi } from "./model/api.ts";

interface MfaRecoveryCodesModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    recoveryCodesRemaining: number;
}

const MfaRecoveryCodesModal = ({ close, isOpen, user, recoveryCodesRemaining }: MfaRecoveryCodesModalProps) => {
    const [code, setCode] = useState("");
    const [error, setError] = useState<string | null>(null);
    const [loading, setLoading] = useState(false);
    const [recoveryCodes, setRecoveryCodes] = useState<string[]>([]);
    const [showCodes, setShowCodes] = useState(false);

    useEffect(() => {
        if (isOpen) {
            setCode("");
            setError(null);
            setRecoveryCodes([]);
            setShowCodes(false);
        }
    }, [isOpen]);

    const handleRegenerate = async () => {
        if (code.length < 6) {
            setError("Please enter a valid verification code");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            const result = await getAccountApi(user).mfaRecoveryCodes({
                body: { code: code }
            });
            setRecoveryCodes(result.recoveryCodes || []);
            setShowCodes(true);
        } catch (e) {
            console.error("Recovery codes regeneration error:", e);
            setError("Failed to regenerate recovery codes. Check your verification code.");
        } finally {
            setLoading(false);
        }
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

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5">
                        <strong>Recovery Codes</strong>
                    </Typography>
                </Grid2>

                {!showCodes ? (
                    <>
                        <Alert severity="info" sx={{ mt: 2, mb: 2 }}>
                            You have <strong>{recoveryCodesRemaining}</strong> recovery codes remaining.
                            {recoveryCodesRemaining <= 3 && (
                                <span> Consider regenerating new codes.</span>
                            )}
                        </Alert>

                        <Typography variant="body1" sx={{ mb: 2 }}>
                            To view or regenerate your recovery codes, enter your current verification code:
                        </Typography>

                        {error && (
                            <Alert severity="error" sx={{ mb: 2 }}>
                                {error}
                            </Alert>
                        )}

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
                            sx={{ mb: 2 }}
                        />

                        <Box sx={{ display: "flex", gap: 1 }}>
                            <Button
                                variant="contained"
                                onClick={handleRegenerate}
                                disabled={loading || code.length < 6}
                            >
                                {loading ? "Regenerating..." : "Regenerate Codes"}
                            </Button>
                            <Button variant="outlined" onClick={close}>
                                Cancel
                            </Button>
                        </Box>
                    </>
                ) : (
                    <>
                        <Alert severity="warning" sx={{ mt: 2, mb: 2 }}>
                            Your previous recovery codes have been invalidated. Save these new codes in a safe place.
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

                        <Button variant="contained" onClick={close}>
                            Done
                        </Button>
                    </>
                )}
            </Box>
        </Modal>
    );
};

export default MfaRecoveryCodesModal;
