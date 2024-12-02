import {Box, Button, Container, TextField, Typography} from "@mui/material";

const Login = () => {
    return (
        <Typography>
            <Container maxWidth="xs">
                <Box sx={{ marginTop: 8, display: "flex", flexDirection: "column", alignItems: "center"}}>
                    <h2>welcome to zrok...</h2>
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
                        />
                        <Button type="submit" fullWidth variant="contained" sx={{ mt: 3, mb: 2 }}>
                            Log In
                        </Button>
                    </Box>
                </Box>
            </Container>
        </Typography>
    );
}

export default Login;