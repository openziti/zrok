import {Box, Button, Checkbox, Container, FormControlLabel, TextField, Typography} from "@mui/material";
import zroket from "./assets/zrok-1.0.0-rocket-purple.svg";
import {useParams} from "react-router";
import {useFormik} from "formik";
import * as Yup from 'yup';
import {useRef, useState} from "react";

const Register = () => {
    const { regToken } = useParams();
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
            console.log(v);
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
        <Typography component="div">
            <Container maxWidth="xs">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zroket} height="300"/>
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
                    <Box component="form" onSubmit={form.handleSubmit}>
                        <Typography component="div" align="center"><h2>Welcome to zrok!</h2></Typography>
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
                        <FormControlLabel control={<Checkbox checked={checked} onChange={toggleChecked} />} label="I accept the Terms of Service" sx={{ mt: 2 }} />
                        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }} disabled={!checked}>
                            Register Account
                        </Button>
                    </Box>
                </Box>
            </Container>
        </Typography>
    );
}

export default Register;