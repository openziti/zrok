import "./styling/index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import {Box, Button, ThemeProvider, Typography} from "@mui/material";
import {theme} from "./styling/theme.ts";
import App from "./App.tsx";
import ErrorBoundary from "./ErrorBoundary.tsx";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <ErrorBoundary fallback={
                <Box sx={{ p: 4, textAlign: "center" }}>
                    <Typography variant="h5" color="error">Something went wrong</Typography>
                    <Button onClick={() => window.location.reload()} variant="outlined" sx={{ mt: 2 }}>Reload</Button>
                </Box>
            }>
                <App />
            </ErrorBoundary>
        </ThemeProvider>
    </StrictMode>
);