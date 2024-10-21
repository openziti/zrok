import "./index.css";
import {StrictMode} from "react";
import {createRoot} from "react-dom/client";
import AgentUi from "./AgentUi.jsx";
import {createTheme, ThemeProvider} from "@mui/material";

export const themeOptions = {
    components: {
        MuiCard: {
            styleOverrides: {
                root: ({theme}) => theme.unstable_sx({
                    mt: 5,
                    p: 2.5,
                    borderRadius: 2,
                }),
            }
        },
        MuiAppBar: {
            styleOverrides: {
                root : ({theme}) => theme.unstable_sx({
                    borderRadius: 2,
                }),
            }
        }
    },
    palette: {
        mode: 'light',
        primary: {
            main: '#241775',
        },
        secondary: {
            main: '#9bf316',
        },
    },
    typography: {
        fontFamily: 'Poppins',
    }
};

createRoot(document.getElementById('root')).render(
  <StrictMode>
      <ThemeProvider theme={createTheme(themeOptions)}>
          <AgentUi />
      </ThemeProvider>
  </StrictMode>,
)
