import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import ApiConsole from "./ApiConsole.tsx";
import {ThemeProvider} from "@mui/material";
import {theme} from "./model/theme.ts";
import {BrowserRouter, Route, Routes} from "react-router";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <BrowserRouter>
                <Routes>
                    <Route path="/" element={<ApiConsole />} />
                </Routes>
            </BrowserRouter>
        </ThemeProvider>
    </StrictMode>
);