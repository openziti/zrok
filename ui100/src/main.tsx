import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import Console from "./Console.tsx";
import {ThemeProvider} from "@mui/material";
import {theme} from "./model/theme.ts";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <Console />
        </ThemeProvider>
    </StrictMode>
);