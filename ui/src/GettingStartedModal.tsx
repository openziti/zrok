import {modalStyle} from "./styling/theme.ts";
import {Box, Grid2, Modal, Typography} from "@mui/material";
import ClipboardText from "./ClipboardText.tsx";
import useApiConsoleStore from "./model/store.ts";

interface GettingStartedModalProps {
    close: () => void;
    isOpen: boolean;
}

const GettingStartedModal = ({ close, isOpen }: GettingStartedModalProps) => {
    const user = useApiConsoleStore(store => store.user);

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Getting Started Quickly</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4>Step 1: Download a zrok Binary</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        The official zrok binaries are published on GitHub:
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <a href="https://github.com/openziti/zrok/releases">https://github.com/openziti/zrok/releases</a>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        <h4>Step 2: Enable Your Operating System Shell</h4>
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
                        <h4>Step 3: Share</h4>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Use the <code>zrok share</code> command to share network connectivity and files (see the
                        <code>--help</code> in the CLI for details: <code>zrok share --help</code>:
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
                        <span> </span>in the <a href="https://docs.zrok.io" target="_">zrok Documentation</a> for more help.
                    </Typography>
                </Grid2>
            </Box>
        </Modal>
    )
}

export default GettingStartedModal;