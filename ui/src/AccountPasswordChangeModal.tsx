import {User} from "./model/user.ts";
import {useEffect, useState} from "react";
import {modalStyle} from "./styling/theme.ts";
import {Box, Button, Grid2, Modal, TextField, Typography} from "@mui/material";
import {useFormik} from "formik";
import * as Yup from 'yup';
import {getAccountApi} from "./model/api.ts";

interface AccountPasswordChangeModalProps {
    close: () => void;
    isOpen: boolean;
    user: User;
}

const AccountPasswordChangeModal =({ close, isOpen, user }: AccountPasswordChangeModalProps) => {
    const [errorMessage, setErrorMessage] = useState<React.JSX.Element>(null);
    const submitButton = <Button color="primary" variant="contained" type="submit" sx={{ mt: 2 }}>Change Password</Button>;
    const [bottomControl, setBottomControl] = useState<React.JSX.Element>(submitButton);

    const passwordChangeForm = useFormik({
        initialValues: {
            currentPassword: "",
            newPassword: "",
            duplicateNewPassword: "",
        },
        onSubmit: v => {
            setErrorMessage(null);
            getAccountApi(user).changePassword({
                body: {
                    email: user.email,
                    oldPassword: v.currentPassword,
                    newPassword: v.newPassword,
                }
            })
                .then(() => {
                    setBottomControl(<Typography>Your password has been changed!</Typography>);
                    setTimeout(() => { close() }, 3000);
                })
                .catch(e => {
                    setErrorMessage(<Typography color="red">Password change failed! Check your current password!</Typography>);
                })
        },
        validationSchema: Yup.object({
            currentPassword: Yup.string()
                .required("Current password is required"),
            newPassword: Yup.string()
                .min(8, "Password must be at least 8 characters")
                .max(64, "Password must be less than 64 characters")
                .matches(
                    /^.*[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?].*$/,
                    "Password requires at least one special character"
                )
                .matches(
                    /^.*[a-z].*$/,
                    "Password requires at least one lowercase letter"
                )
                .matches(
                    /^.*[A-Z].*$/,
                    "Password requires at least one uppercase letter"
                )
                .required("Password is required"),
            duplicateNewPassword: Yup.string()
                .required("Please confirm your new password")
                .test("password-matches", "Password confirmation does not match", v => v === passwordChangeForm.values.newPassword)
        }),
    });

    useEffect(() => {
        passwordChangeForm.values.currentPassword = "";
        passwordChangeForm.values.newPassword = "";
        passwordChangeForm.values.duplicateNewPassword = "";
        setErrorMessage(null);
        setBottomControl(submitButton);
    }, [isOpen]);

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
                            error={passwordChangeForm.errors.currentPassword !== undefined}
                            helperText={passwordChangeForm.errors.currentPassword}
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
                            error={passwordChangeForm.errors.newPassword !== undefined}
                            helperText={passwordChangeForm.errors.newPassword}
                            sx={{ mt: 2 }}
                        />
                        <TextField
                            fullWidth
                            type="password"
                            id="duplicateNewPassword"
                            name="duplicateNewPassword"
                            label="Confirm New Password"
                            value={passwordChangeForm.values.duplicateNewPassword}
                            onChange={passwordChangeForm.handleChange}
                            onBlur={passwordChangeForm.handleBlur}
                            error={passwordChangeForm.errors.duplicateNewPassword !== undefined}
                            helperText={passwordChangeForm.errors.duplicateNewPassword}
                            sx={{ mt: 2 }}
                        />
                    </Grid2>
                    { errorMessage ? <Grid2 container sx={{ mt: 2, p: 1}}><Typography>{errorMessage}</Typography></Grid2> : null}
                    <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                        {bottomControl}
                    </Grid2>
                </form>
            </Box>
        </Modal>
    );
}

export default AccountPasswordChangeModal;