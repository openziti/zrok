import {User} from "./model/user.ts";
import {useEffect, useState} from "react";
import {modalStyle} from "./styling/theme.ts";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Typography} from "@mui/material";
import {getAccountApi} from "./model/api.ts";
import useApiConsoleStore from "./model/store.ts";
import ClipboardText from "./ClipboardText.tsx";
import {extractErrorMessage} from "./model/errors.ts";

interface RegenerateAccountTokenModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
}

type RegenerateTokenStatus = "idle" | "success" | "failed";

const RegenerateAccountTokenModal = ({ close, isOpen, user }: RegenerateAccountTokenModalProps) => {
    const updateUser = useApiConsoleStore((state) => state.updateUser);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);
    const [newToken, setNewToken] = useState<string | null>(null);
    const [status, setStatus] = useState<RegenerateTokenStatus>("idle");
    const [checked, setChecked] = useState<boolean>(false);

    const reload = () => {
        window.location.reload();
    }

    useEffect(() => {
        setChecked(false);
        setErrorMessage(null);
        setNewToken(null);
        setStatus("idle");
    }, [isOpen]);

    const regenerateToken = () => {
        getAccountApi(user).regenerateAccountToken({body: {emailAddress: user.email}})
            .then(d => {
                const newUser = {
                    email: user.email!,
                    token: d.accountToken!,
                }
                updateUser(newUser);
                localStorage.setItem("user", JSON.stringify(newUser));
                document.dispatchEvent(new Event("userUpdated"));
                setErrorMessage(null);
                setNewToken(d.accountToken ?? null);
                setStatus("success");
            })
            .catch(async (e) => {
                const msg = await extractErrorMessage(e, "failed to regenerate account token");
                setErrorMessage(msg);
                setNewToken(null);
                setStatus("failed");
            });
    }

    const controls = <>
        <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
        <FormControlLabel control={<Checkbox checked={checked} onChange={(_, nextChecked) => setChecked(nextChecked)} />} label={<p>I confirm that I want to regenerate my account token</p>} sx={{ mt: 2 }} />
        </Grid2>
        <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
            <Button color="error" variant="contained" disabled={!checked} onClick={regenerateToken}>Regenerate Account Token</Button>
        </Grid2>
    </>;

    return (
        <Modal open={isOpen} onClose={close} aria-labelledby="modal-title-regenerate-token">
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5" id="modal-title-regenerate-token"><strong>Regenerate Account Token</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h6" color="red">
                        WARNING: Regenerating your account token can stop all environments and shares from operating properly!
                        Please read the following instructions to prevent interruptions!
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        You will need to use the <code> zrok rebase accountToken </code> command to update any enabled
                        environments to use your new account token. Rebasing your environments will minimize any service
                        disruptions caused by regenerating your account token.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Keep in mind that once you've regenerated your account token, any running <code> zrok share </code>
                        or <code> zrok access </code> processes may not be able to interact with the zrok service properly
                        until they are restarted.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Alternatively, you can just <code> zrok disable </code> any enabled environments and re-enable
                        using the updated account token. Running <code> zrok disable </code> before you regenerate will
                        delete your environments and any shares they contain (including reserved shares). So if you have
                        environments and reserved shares you need to preserve, your best option is to use the
                        <code> zrok rebase accountToken </code> command as described above.
                    </Typography>
                </Grid2>
                {status !== "success" ? controls : null}
                {status === "success" && newToken ? <>
                    <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                        <Typography variant="h6" sx={{ mt: 2, p: 1 }}>
                            Your new account token is: <code>{newToken}</code> <ClipboardText text={newToken} />
                        </Typography>
                    </Grid2>
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Button type="primary" variant="contained" onClick={reload}>Reload API Console</Button>
                    </Grid2>
                </> : null}
                {errorMessage ? <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Typography color="error">{errorMessage}</Typography>
                </Grid2> : null}
            </Box>
        </Modal>
    );
}

export default RegenerateAccountTokenModal;
