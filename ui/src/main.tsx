import "./styling/index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import {ThemeProvider} from "@mui/material";
import {theme} from "./styling/theme.ts";
import App from "./App.tsx";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <App />
        </ThemeProvider>
    </StrictMode>
);