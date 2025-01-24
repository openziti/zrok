import {User} from "./model/user.ts";
import {useState} from "react";
import {modalStyle} from "./styling/theme.ts";
import {Box, Button, Grid2, Modal, TextField, Typography} from "@mui/material";
import {useFormik} from "formik";

interface AccountPasswordChangeModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
}

const AccountPasswordChangeModal =({ close, isOpen, user }: AccountPasswordChangeModalProps) => {
    const [errorMessage, setErrorMessage] = useState<React.JSX.Element>(null);

    const passwordChangeForm = useFormik({
        initialValues: {
            currentPassword: "",
            newPassword: "",
            duplicateNewPassword: "",
        },
        onSubmit: v => {
            setErrorMessage(null as React.JSX.Element);
            // api
        }
    });

    return (
        <Modal open={isOpen} onClose={close}>
            <Box sx={{ ...modalStyle }}>
                <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                    <Typography variant="h5"><strong>Change Password</strong></Typography>
                </Grid2>
                <form onSubmit={passwordChangeForm.handleSubmit}>
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <TextField
                            fullWidth
                            type="password"
                            id="currentPassword"
                            name="currentPassword"
                            label="Current Password"
                            value={passwordChangeForm.values.currentPassword}
                            onChange={passwordChangeForm.handleChange}
                            onBlur={passwordChangeForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                        <TextField
                            fullWidth
                            type="password"
                            id="newPassword"
                            name="newPassword"
                            label="New Password"
                            value={passwordChangeForm.values.newPassword}
                            onChange={passwordChangeForm.handleChange}
                            onBlur={passwordChangeForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                        <TextField
                            fullWidth
                            type="password"
                            id="duplicateNewPassword"
                            name="duplicateNewPassword"
                            label="Re-enter New Password"
                            value={passwordChangeForm.values.duplicateNewPassword}
                            onChange={passwordChangeForm.handleChange}
                            onBlur={passwordChangeForm.handleBlur}
                            sx={{ mt: 2 }}
                        />
                    </Grid2>
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Change Password</Button>
                    </Grid2>
                </form>
            </Box>
        </Modal>
    );
}

export default AccountPasswordChangeModal;