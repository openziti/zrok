import {Box, Button, Container, TextField, Typography} from "@mui/material";
import {useState} from "react";
import zroket from "./assets/zrok-1.0.0-rocket-purple.svg";

const ForgotPassword = () => {
    const [email, setEmail] = useState("");

    return (
        <Typography component="div">
            <Container maxWidth="xs">
                <Box sx={{marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <img src={zroket} height="300"/>
                    <h1 style={{ color: "#241775" }}>z r o k</h1>
                    <Box component="form" noValidate>
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
                            onChange={v => { setEmail(v.target.value) }}
                        />
                        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }} style={{ color: "#9bf316" }}>
                            Send Password Reset Email
                        </Button>
                    </Box>
                </Box>
            </Container>
        </Typography>
    );
}

export default ForgotPassword;