import {Box, Button, Container, Paper, TextField, Typography} from "@mui/material";
import zrokLogo from "./assets/zrok-1.0.0-rocket-purple.svg";
import {Link} from "react-router";
import {AccountApi} from "./api";
import {useFormik} from "formik";
import * as Yup from 'yup';
import {useState} from "react";

interface ForgotPasswordFormProps {
    doRequest: ({ email: string }) => void;
}

const ForgotPasswordForm = ({ doRequest }: ForgotPasswordFormProps) => {
    const form = useFormik({
        initialValues: {
            email: ""
        },
        onSubmit: v => {
            console.log(v);
            doRequest(v.email);
        },
        validationSchema: Yup.object({
            email: Yup.string().email()
        })
    });

    return (
        <form onSubmit={form.handleSubmit}>
            <Typography component="div" align="center"><h2>Forgot Your Password?</h2></Typography>
            <TextField
                fullWidth
                id="email"
                name="email"
                label="Email Address"
                autoFocus
                value={form.values.email}
                onChange={form.handleChange}
                onBlur={form.handleBlur}
                error={form.errors.email !== undefined}
                helperText={form.errors.email}
            />
            <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }}>
                Send Password Reset Email
            </Button>
            <Box component="div" style={{ textAlign: "center" }}>
                <Link to="/">Return to Login</Link>
            </Box>
        </form>
    );
}

const RequestSubmittedMessage = () => {
    return (
        <Paper sx={{ p: 5 }}>
            <Box component="div">
                <Typography component="div" align="center"><h2>Request Submitted...</h2></Typography>
                <Typography component="div">
                    <p>
                        If your email address is found, you will be sent an email with a link to reset your password.
                    </p>
                </Typography>
                <Typography component="div">
                    <p>
                        <strong>Please check your "spam" folder for this email if you do not receive it after a few minutes!</strong>
                    </p>
                </Typography>
                <Box component="div" style={{ textAlign: "center" }}>
                    <Link to="/">Return to Login</Link>
                </Box>
            </Box>
        </Paper>
    );
}

const ForgotPassword = () => {
    const requestResetPassword = (email) => {
        console.log("requestResetPassword", email);
        new AccountApi().resetPasswordRequest({body: {emailAddress: email}})
            .then(() => {
                setControl(<RequestSubmittedMessage />);
            })
            .catch(e => {
                console.log("resetPasswordRequest", e);
            })
    }

    const [control, setControl] = useState<React.JSX.Element>(<ForgotPasswordForm doRequest={requestResetPassword} />);

    return (
        <Typography component="div">
            <Container maxWidth="xs">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zrokLogo} height={300} />
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
                    {control}
                </Box>
            </Container>
        </Typography>
    );
}

export default ForgotPassword;