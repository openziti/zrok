import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import AgentUi from "./AgentUi.jsx";
import {createTheme, ThemeProvider} from "@mui/material";
import {themeOptions} from "./model/theme.js";

createRoot(document.getElementById('root')).render(
  <StrictMode>
      <ThemeProvider theme={createTheme(themeOptions)}>
          <AgentUi />
      </ThemeProvider>
  </StrictMode>,
);
