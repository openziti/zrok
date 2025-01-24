import {User} from "./model/user.ts";
import {useEffect, useRef, useState} from "react";
import {modalStyle} from "./styling/theme.ts";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Typography} from "@mui/material";

interface RegenerateAccountTokenModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
}

const RegenerateAccountTokenModal = ({ close, isOpen, user }: RegenerateAccountTokenModalProps) => {
    const [errorMessage, setErrorMessage] = useState<React.JSX.Element>(null);
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>(checked);

    const toggleChecked = () => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
    }, [isOpen]);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Regenerate Account Token</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h6" color="red">
                        WARNING: Regenerating your account token will stop all environments and shares from operating properly!
                        Please read the following instructions to prevent interruptions!
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        You will need to manually edit your <code>$&#123;HOME&#125;/.zrok/environment.json</code> files
                        (in each environment) to use the new <code>zrok_token</code>. Updating these files will restore
                        the functionality of your environments.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Alternatively, you can just <code>zrok disable</code> any enabled environments and re-enable
                        using the updated account token. Running <code>zrok disable</code> before you regenerate will
                        delete your environments and any shares they contain (including reserved shares). So if you have
                        environments and reserved shares you need to preserve, your best option is to update the
                        <code>zrok_token</code> in those environments as described above.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm that I want to regenerate my account token</p>} sx={{ mt: 2 }} />
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked}>Regenerate Account Token</Button>
                </Grid2>
            </Box>
        </Modal>
    );
}

export default RegenerateAccountTokenModal;