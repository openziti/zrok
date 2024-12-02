import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import ApiConsole from "./ApiConsole.tsx";
import {ThemeProvider} from "@mui/material";
import {theme} from "./model/theme.ts";
import {BrowserRouter, Route, Routes} from "react-router";
import Login from "./Login.tsx";

createRoot(document.getElementById('root')!).render(
    <StrictMode>
        <ThemeProvider theme={theme}>
            <BrowserRouter>
                <Routes>
                    <Route index element={<ApiConsole />} />
                    <Route path="login" element={<Login />} />
                </Routes>
            </BrowserRouter>
        </ThemeProvider>
    </StrictMode>
);