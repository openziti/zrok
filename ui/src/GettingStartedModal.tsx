import {modalStyle} from "./styling/theme.ts";
import {Box, Grid2, Modal, Typography} from "@mui/material";

interface GettingStartedModalProps {
    close: () => void;
    isOpen: boolean;
}

const GettingStartedModal = ({ close, isOpen }: GettingStartedModalProps) => {
    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Getting Started with zrok</strong></Typography>
                </Grid2>
            </Box>
        </Modal>
    )
}

export default GettingStartedModal;