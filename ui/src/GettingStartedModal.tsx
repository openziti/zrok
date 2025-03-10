import {modalStyle} from "./styling/theme.ts";
import {Box, Grid2, Modal, Typography} from "@mui/material";
import ClipboardText from "./ClipboardText.tsx";
import useApiConsoleStore from "./model/store.ts";
import {CSSProperties, useEffect, useState} from "react";
import CheckMarkIcon from "@mui/icons-material/Check";

interface GettingStartedModalProps {
    close: () => void;
    isOpen: boolean;
}

const GettingStartedModal = ({ close, isOpen }: GettingStartedModalProps) => {
    const user = useApiConsoleStore(store => store.user);
    const nodes = useApiConsoleStore(store => store.nodes);
    const [environmentCreated, setEnvironmentCreated] = useState<boolean>(false);
    const [shareCreated, setShareCreated] = useState<boolean>(false);
    const checkedStyle: CSSProperties = { color: "green" };
    const uncheckedStyle: CSSProperties = null as CSSProperties;

    const inspectState = () => {
        let environments = 0;
        let shares = 0;
        nodes.forEach(node => {
            if(node.type === "environment") {
                environments++;
            }
            if(node.type === "share") {
                shares++;
            }
        });
        setEnvironmentCreated(environments > 0);
        setShareCreated(shares > 0);
    }

    useEffect(() => {
        inspectState();
    }, []);

    useEffect(() => {
        inspectState();
    }, [nodes]);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Getting Started Wizard</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4 style={ environmentCreated ? checkedStyle : uncheckedStyle }>Step 1: Get a zrok Binary { environmentCreated ? <CheckMarkIcon/> : null }</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        The documentation has a guide for downloading the right zrok binary for your system:
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <a href="https://docs.zrok.io/docs/getting-started#installing-the-zrok-command" target="_">https://docs.zrok.io/docs/getting-started#installing-the-zrok-command</a>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4 style={ environmentCreated ? checkedStyle : uncheckedStyle }>Step 2: Enable Your Operating System Shell { environmentCreated ? <CheckMarkIcon/> : null }</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Create a zrok "environment" by using the <code>zrok enable</code> command:
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <pre>
                        $ zrok enable {user.token} <ClipboardText text={"zrok enable " + user.token}/>
                        </pre>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4 style={ shareCreated ? checkedStyle : uncheckedStyle }>Step 3: Share { shareCreated ? <CheckMarkIcon/> : null }</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Use the <code>zrok share</code> command to share network connectivity and files (see the
                        <code> --help</code> in the CLI for details: <code>zrok share --help</code>):
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <pre>
                        $ zrok share public --backend-mode web . <ClipboardText text={"zrok share public --backend-mode web ."} />
                        </pre>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Share the generated URL, allowing secure access to the current directory.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4>Need More Help?</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Visit the <a href="https://docs.zrok.io/docs/getting-started" target="_">Getting Started Guide</a>
                        <span> </span>and the <a href="https://docs.zrok.io" target="_">zrok Documentation</a> for more help.
                    </Typography>
                </Grid2>
            </Box>
        </Modal>
    )
}

export default GettingStartedModal;