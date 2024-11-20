import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import ApiConsole from "./ApiConsole.tsx";
import {ThemeProvider} from "@mui/material";
import {theme} from "./model/theme.ts";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <ApiConsole />
        </ThemeProvider>
    </StrictMode>
);