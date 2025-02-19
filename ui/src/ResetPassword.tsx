import {useEffect, useState} from "react";
import {Link, useParams} from "react-router";
import {useFormik} from "formik";
import * as Yup from "yup";
import {Box, Button, Container, TextField, Typography} from "@mui/material";
import zrokLogo from "./assets/zrok-1.0.0-rocket-purple.svg";
import {AccountApi} from "./api";

interface ResetPasswordFormProps {
    resetPassword: (v) => void;
}

const ResetPasswordForm = ({ resetPassword }: ResetPasswordFormProps) => {
    const form = useFormik({
        initialValues: {
            password: "",
            confirm: "",
        },
        onSubmit: v => {
            resetPassword(v);
        },
        validationSchema: Yup.object({
            password: Yup.string()
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
            confirm: Yup.string()
                .required("Please confirm your new password")
                .test("password-matches", "Password confirmation does not match", v => v === form.values.password)
        })
    });

    return (
        <form onSubmit={form.handleSubmit}>
            <Typography component="div" align="center"><h2>New Password</h2></Typography>
            <TextField
                fullWidth
                type="password"
                id="password"
                name="password"
                label="Create a Password"
                value={form.values.password}
                onChange={form.handleChange}
                onBlur={form.handleBlur}
                error={form.errors.password !== undefined}
                helperText={form.errors.password}
                sx={{ mt: 2 }}
            />
            <TextField
                fullWidth
                type="password"
                id="confirm"
                name="confirm"
                label="Confirm Your Password"
                value={form.values.confirm}
                onChange={form.handleChange}
                onBlur={form.handleBlur}
                error={form.errors.confirm !== undefined}
                helperText={form.errors.confirm}
                sx={{ mt: 2 }}
            />
            <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }}>
                Set New Password
            </Button>
        </form>
    );
}

const ResetComplete = () => {
    return (
        <Box component="div">
            <Container>
                <Box sx={{ display: "flex", alignItems: "center" }}>
                    <Typography component="div">
                        <h2>Password Changed...</h2>
                    </Typography>
                </Box>
            </Container>
            <Container>
                <Box sx={{ mt: 3 }}>
                    <Typography component="p">
                        You can use your new password to log into the console here:
                    </Typography>
                </Box>
                <Box sx={{ mt: 3 }}>
                    <Typography component="p">
                        <a href={window.location.origin}>{window.location.origin}</a>
                    </Typography>
                </Box>
            </Container>
        </Box>
    );
}

const ResetFailed = () => {
    return (
        <Box component="div">
            <Container>
                <Box sx={{ display: "flex", alignItems: "center" }} >
                    <Typography component="div">
                        <h2 style={{ color: "red" }}>Password Reset Failed!</h2>
                    </Typography>
                </Box>
            </Container>
            <Container>
                <Box sx={{ mt: 3 }}>
                    <Typography component="p">
                        Your password change request failed. This might be due to an invalid password reset token.
                    </Typography>
                </Box>
                <Box sx={{ mt: 3 }}>
                    <Typography component="p">
                        Please use the forgot password link below to create a new password reset request:
                    </Typography>
                </Box>
                <Box sx={{ mt: 3 }}>
                    <Typography component="p">
                        <Link to="/forgotPassword">Forgot Password?</Link>
                    </Typography>
                </Box>
            </Container>
        </Box>
    );
}

const ResetPassword = () => {
    const { resetToken } = useParams();
    const [component, setComponent] = useState<React.JSX.Element>();

    const doReset = (v) => {
        new AccountApi().resetPassword({body: {resetToken: resetToken, password: v.password}})
            .then(() => {
                setComponent(<ResetComplete />);
            })
            .catch(e => {
                setComponent(<ResetFailed />);
                console.log("doReset", e);
            });
    }

    useEffect(() => {
        setComponent(<ResetPasswordForm resetPassword={doReset} />);
    }, [resetToken]);

    return (
        <Typography component="div">
            <Container maxWidth="sm">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zrokLogo} height="300"/>
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
                    {component}
                </Box>
            </Container>
        </Typography>
    )
}

export default ResetPassword;