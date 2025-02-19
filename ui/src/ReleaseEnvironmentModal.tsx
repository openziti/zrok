import {Environment} from "./api";
import {useEffect, useRef, useState} from "react";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Typography} from "@mui/material";
import {modalStyle} from "./styling/theme.ts";
import {User} from "./model/user.ts";
import {Node} from "@xyflow/react";
import {getEnvironmentApi} from "./model/api.ts";

interface ReleaseEnvironmentProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    environment: Node;
    detail: Environment;
}

const ReleaseEnvironmentModal = ({ close, isOpen, user, environment, detail }: ReleaseEnvironmentProps) => {
    const [errorMessage, setErrorMessage] = useState<React.JSX.Element>(null);
    const [description, setDescription] = useState<String>("");
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>();
    checkedRef.current = checked;

    const toggleChecked = () => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
        setErrorMessage(null);
    }, [isOpen]);

    useEffect(() => {
        if(detail && detail.description) {
            setDescription(detail.description);
        }
    }, [detail]);

    const releaseEnvironment = () => {
        if(environment.data && environment.data.envZId) {
            getEnvironmentApi(user).disable({
                body: {
                    identity: environment.data.envZId as string
                }
            })
                .then(() => {
                    close();
                })
                .catch(e => {
                    e.response.json().then(ex => {
                        console.log("releaseEnvironment", ex.message);
                    });
                    setErrorMessage(<Typography color="red">An error occurred releasing your environment <code>{environment.id}</code>!</Typography>);
                    setTimeout(() => {
                        setErrorMessage(null);
                        setChecked(false);
                    }, 2000);
                });
        }
    }

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Release Environment</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Would you like to release the environment <code>{description}</code> ?</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Releasing this environment will also release any shares and accesses that are associated with it.</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm the release of <code>{description}</code></p>} sx={{ mt: 2 }} />
                </Grid2>
                { errorMessage ? <Grid2 container sx={{ mb: 2, p: 1}}><Typography>{errorMessage}</Typography></Grid2> : null}
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked} onClick={releaseEnvironment}>Release</Button>
                </Grid2>
            </Box>
        </Modal>
    );
}

export default ReleaseEnvironmentModal;