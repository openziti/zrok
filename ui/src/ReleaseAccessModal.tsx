import {Frontend} from "./api";
import {useEffect, useRef, useState} from "react";
import {Box, Button, Checkbox, FormControlLabel, Grid2, Modal, Typography} from "@mui/material";
import {modalStyle} from "./styling/theme.ts";
import {User} from "./model/user.ts";
import {Node} from "@xyflow/react";
import {getShareApi} from "./model/api.ts";

interface ReleaseAccessProps {
    close: () => void;
    isOpen: boolean;
    user: User;
    access: Node;
    detail: Frontend;
}

const ReleaseAccessModal = ({ close, isOpen, user, access, detail }: ReleaseAccessProps) => {
    const [errorMessage, setErrorMessage] = useState<React.JSX.Element>(null);
    const [frontendToken, setFrontendToken] = useState<String>("");
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>(checked);

    const toggleChecked = () => {
        setChecked(!checkedRef.current);
    }

    useEffect(() => {
        setChecked(false);
        setErrorMessage(null);
    }, [isOpen]);

    useEffect(() => {
        if(detail && detail.frontendToken) {
            setFrontendToken(detail.frontendToken);
        }
    }, [detail]);

    const releaseAccess = () => {
        setErrorMessage(null);
        if(detail && detail.frontendToken) {
            getShareApi(user).unaccess({
                body: {
                    frontendToken: detail.frontendToken,
                    envZId: access.data.envZId as string,
                    shareToken: detail.shareToken
                }
            })
                .then(() => {
                    close();
                })
                .catch(e => {
                    e.response.json().then(ex => {
                        console.log("releaseAccess", ex.message);
                    });
                    setErrorMessage(<Typography color="red">An error occurred releasing your access <code>{detail.frontendToken}</code>!</Typography>);
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
                    <Typography variant="h5"><strong>Release Access</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="body1">Would you like to release the access <code>{frontendToken}</code> ?</Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I confirm the release of <code>{frontendToken}</code></p>} sx={{ mt: 2 }} />
                </Grid2>
                { errorMessage ? <Grid2 container sx={{ mb: 2, p: 1}}><Typography>{errorMessage}</Typography></Grid2> : null}
                <Grid2 container sx={{ flexGrow: 1 }} alignItems="center">
                    <Button color="error" variant="contained" disabled={!checked} onClick={releaseAccess}>Release</Button>
                </Grid2>
            </Box>
        </Modal>
    );
}

export default ReleaseAccessModal;