import {createTheme} from "@mui/material";

export const COLORS = {
    primary: '#241775',
    secondary: '#9bf316',
    metrics: '#04adef',
    alertBannerBg: '#f5fde7',
} as const;

const componentOptions = {
    MuiCard: {
        styleOverrides: {
            root: ({theme}) => theme.unstable_sx({
                mt: 5,
                p: 1,
                borderRadius: 3,
            }),
        }
    },
    MuiAppBar: {
        styleOverrides: {
            root : ({theme}) => theme.unstable_sx({
                borderRadius: 3,
            }),
        }
    }
}

export const theme = createTheme({
    components: componentOptions,
    palette: {
        mode: 'light',
        primary: {
            main: COLORS.primary,
        },
        secondary: {
            main: COLORS.secondary,
        },
    },
    typography: {
        fontFamily: 'Poppins',
    },
})

export const modalStyle = {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: 600,
    bgcolor: 'background.paper',
    boxShadow: 24,
    p: 4,
    borderRadius: 2,
};