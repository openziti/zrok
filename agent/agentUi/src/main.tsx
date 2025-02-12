import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import {ThemeProvider} from "@mui/material";
import {theme} from "./model/theme.ts";
import AgentUi from "./AgentUi.tsx";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <AgentUi />
        </ThemeProvider>
    </StrictMode>
);