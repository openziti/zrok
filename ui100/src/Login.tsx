import {Box, Button, Container, TextField, Typography} from "@mui/material";
import {User} from "./model/user.ts";
import {useEffect, useState} from "react";
import {AccountApi, MetadataApi} from "./api";
import {Link} from "react-router";
import zroket from "./assets/zrok-1.0.0-rocket-purple.svg";

interface LoginProps {
    onLogin: (user: User) => void;
}

const Login = ({ onLogin }: LoginProps) => {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [message, setMessage] = useState("");
    const [tou, setTou] = useState(null as string);

    useEffect(() => {
        new MetadataApi()._configuration()
            .then(d => {
                if(d.touLink && d.touLink.trim() !== "") {
                    setTou(d.touLink);
                }
            })
            .catch(e => {
                console.log(e);
            });
    }, []);

    const login = async e => {
        e.preventDefault();

        new AccountApi().login({body: {"email": email, "password": password}})
            .then(d => {
                onLogin({email: email, token: d.toString()});
            })
            .catch(e => {
                console.log(e);
                setMessage("login failed!")
            });
    }

    return (
        <Typography component="div">
            <Container maxWidth="xs">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zroket} height="300"/>
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
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
                            onChange={v => {
                                setMessage("");
                                setEmail(v.target.value)
                            }}
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
                            onChange={v => {
                                setMessage("");
                                setPassword(v.target.value)
                            }}
                        />
                        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }}>
                            Log In
                        </Button>
                        <Box component="div" style={{ textAlign: "center" }}>
                            <Box component="h3" style={{ color: "red" }}>{message}</Box>
                        </Box>
                        <Box component="div" style={{ textAlign: "center" }}>
                            <Link to="/forgotPassword">Forgot Password?</Link>
                        </Box>
                        <Box component="div" style={{ textAlign: "center" }}>
                            <div dangerouslySetInnerHTML={{__html: tou}}></div>
                        </Box>
                    </Box>
                </Box>
            </Container>
        </Typography>
    );
}

export default Login;