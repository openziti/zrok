import {Box, Button, Checkbox, CircularProgress, Container, FormControlLabel, Grid2, Paper, TextField, Typography} from "@mui/material";
import zrokLogo from "./assets/zrok-1.0.0-rocket-purple.svg";
import {useParams} from "react-router";
import {useFormik} from "formik";
import * as Yup from 'yup';
import {useEffect, useRef, useState} from "react";
import {AccountApi, MetadataApi} from "./api";
import ClipboardText from "./ClipboardText.tsx";
import {sanitizeHtml} from "./model/html.ts";
import {extractErrorMessage} from "./model/errors.ts";

interface SetPasswordValues {
    password: string;
    confirm: string;
}

interface SetPasswordFormProps {
    email: string;
    touLink: string;
    register: (v: SetPasswordValues) => void;
}

const SetPasswordForm = ({ email, touLink, register }: SetPasswordFormProps) => {
    const [checked, setChecked] = useState<boolean>(false);
    const checkedRef = useRef<boolean>();
    checkedRef.current = checked;
    const toggleChecked = () => { setChecked(!checkedRef.current) }

    const form = useFormik({
        initialValues: {
            password: "",
            confirm: "",
        },
        onSubmit: v => {
            register(v);
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
        <Box component="form" onSubmit={form.handleSubmit}>
            <Typography component="div" align="center"><h2>Welcome to zrok!</h2></Typography>
            <Typography component="p" align="center" sx={{ mb: 2 }}><code>{email}</code></Typography>
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
            <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label={<p>I accept the <span dangerouslySetInnerHTML={{__html: sanitizeHtml(touLink ?? "")}}></span></p>} sx={{ mt: 2 }} />
            <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }} disabled={!checked}>
                Register Account
            </Button>
        </Box>
    );
}

interface RegistrationCompleteProps {
    token: string;
}

const RegistrationComplete = ({ token }: RegistrationCompleteProps) => {
    return (
        <Paper sx={{ p: 5 }}>
            <Box component="div">
                <Container>
                    <Box sx={{ display: "flex", alignItems: "center" }}>
                        <Typography component="div">
                            <h2>Registration completed!</h2>
                        </Typography>
                    </Box>
                </Container>
                <Container>
                    <Box>
                        <Typography component="p">
                            Your account was created successfully!
                        </Typography>
                    </Box>
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            Your new account token is: <code>{token}</code> <ClipboardText text={token} />
                        </Typography>
                    </Box>
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            You can create an environment using your account token, like this:
                        </Typography>
                    </Box>
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            <code>$ zrok enable {token}</code> <ClipboardText text={"zrok enable " + token} />
                        </Typography>
                    </Box>
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            <strong>Your account token is a secret (like a password).
                                Please do not share it with anyone!</strong>
                        </Typography>
                    </Box>
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
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            <h3>Enjoy zrok!</h3>
                        </Typography>
                    </Box>
                </Container>
            </Box>
        </Paper>
    );
}

const InvalidToken = () => {
    return (
        <Paper sx={{ p: 5 }}>
            <Box component="div">
                <Container>
                    <Box sx={{ display: "flex", alignItems: "center" }}>
                        <Typography component="div"><h2 style={{ color: "red" }} align="center">Invalid registration token?!</h2></Typography>
                    </Box>
                </Container>
                <Container>
                    <Box>
                        <Typography component="p">
                            Your registration token may have expired!
                        </Typography>
                    </Box>
                    <Box sx={{ mt: 3 }}>
                        <Typography component="p">
                            Please use the <code>zrok invite</code> command to
                            generate a new registration request and try again.
                        </Typography>
                    </Box>
                </Container>
            </Box>
        </Paper>
    );
}

const Register = () => {
    const { regToken } = useParams();
    const [component, setComponent] = useState<React.JSX.Element>(
        <Grid2 container sx={{ justifyContent: "center", mt: 4 }}><CircularProgress /></Grid2>
    );
    const [error, setError] = useState<boolean>(false);
    const [errorMessage, setErrorMessage] = useState<string>("");
    const [email, setEmail] = useState<string>();
    const [touLink, setTouLink] = useState<string>();

    const doRegistration = (v: SetPasswordValues) => {
        setErrorMessage("");
        new AccountApi().register({body: {registerToken: regToken, password: v.password}})
            .then(d => {
                setComponent(<RegistrationComplete token={d.accountToken!} />);
            })
            .catch(async (e) => {
                const msg = await extractErrorMessage(e, "registration failed");
                setErrorMessage(msg);
            });
    }

    useEffect(() => {
        if(regToken) {
            const controller = new AbortController();
            new AccountApi().verify({body: {registerToken: regToken}}, { signal: controller.signal })
                .then((d) => {
                    setEmail(d.email);
                })
                .catch(() => {
                    setError(true);
                });
            return () => controller.abort();
        }
    }, [regToken]);

    useEffect(() => {
        if(email) {
            const controller = new AbortController();
            new MetadataApi()._configuration({ signal: controller.signal })
                .then(d => {
                    setTouLink(d.touLink);
                })
                .catch(() => {});
            return () => controller.abort();
        }
    }, [email]);

    useEffect(() => {
        if(!error && email && touLink) {
            setComponent(<SetPasswordForm email={email!} touLink={touLink!} register={doRegistration} />);
        } else {
            if(error) {
                setComponent(<InvalidToken />);
            }
        }
    }, [touLink, error]);

    return (
        <Typography component="div">
            <Container maxWidth="sm">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zrokLogo} height="300" alt="zrok logo"/>
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
                    { errorMessage && <Typography color="error">{errorMessage}</Typography> }
                    {component}
                </Box>
            </Container>
        </Typography>
    );
}

export default Register;