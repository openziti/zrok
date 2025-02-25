import {modalStyle} from "./styling/theme.ts";
import {Box, Grid2, Modal, Typography} from "@mui/material";

interface BandwidthLimitedModalProps {
    close: () => void;
    isOpen: boolean;
}

const BandwidthLimitedModal = ({ close, isOpen }: BandwidthLimitedModalProps) => {
    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Bandwidth Limit Exceeded!</strong></Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h6" color="red">
                        Your zrok account has exceeded the configured bandwidth limit!
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Your zrok account is configured with a data transfer (bandwidth) limit that describes the amount
                        of data you can send and receive on the network within a specified period of time. As of now,
                        your account is currently over this configured limit.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        When your account is in this limited state, your shares will temporarily be disabled, and users
                        of your shares will be unable to access them. When limited, you cannot create any additional
                        shares, accesses, or environments.
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        To remove the limit, you can either:
                        <ul>
                            <li>Allow enough time to pass, such that your per-period data transfer falls back below
                            the configured threshold</li>
                            <li>Upgrade your account to have a higher bandwidth limit</li>
                        </ul>
                    </Typography>
                </Grid2>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography>
                        Once the limit expires or is otherwise removed, your shares and accesses will resume working
                        normally.
                    </Typography>
                </Grid2>
            </Box>
        </Modal>
    );
}

export default BandwidthLimitedModal;