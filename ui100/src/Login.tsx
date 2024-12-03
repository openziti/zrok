import {Box, Button, Container, TextField, Typography} from "@mui/material";
import {User} from "./model/user.ts";
import {useState} from "react";
import {AccountApi} from "./api";

interface LoginProps {
    onLogin: (user: User) => void;
}

const Login = ({ onLogin }: LoginProps) => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [message, setMessage] = useState("");

    const login = async e => {
        e.preventDefault();
        console.log(email, password);

        let api = new AccountApi();
        api.login({body: {"email": email, "password": password}})
            .then(d => {
                onLogin({email: email, token: d.toString()});
            })
            .catch(e => {
                setMessage("login failed: " + e.toString());
            });
    }

    return (
        <Typography component="div">
            <Container maxWidth="xs">
                <Box sx={{ marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <h2>welcome to zrok...</h2>
                    <Box component="form" noValidate onSubmit={login}>
                        <TextField
                            margin="normal"
                            required
                            fullWidth
                            id="email"
                            label="Email Address"
                            name="email"
                            autoComplete="email"
                            autoFocus
                            value={email}
                            onChange={v => { setMessage(""); setEmail(v.target.value) }}
                        />
                        <TextField
                            margin="normal"
                            required
                            fullWidth
                            name="password"
                            label="Password"
                            type="password"
                            id="password"
                            autoComplete="current-password"
                            value={password}
                            onChange={v => { setMessage(""); setPassword(v.target.value) }}
                        />
                        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
                            Log In
                        </Button>
                        <Box component="h3" style={{ color: "red" }}>{message}</Box>
                    </Box>
                </Box>
            </Container>
        </Typography>
    );
}

export default Login;