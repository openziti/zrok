import {createTheme} from "@mui/material";

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
            main: '#241775',
        },
        secondary: {
            main: '#9bf316',
        },
    },
    typography: {
        fontFamily: 'Poppins',
    },
})

export const modalStyle = {
    position: 'absolute',
    top: '25%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    width: 600,
    bgcolor: 'background.paper',
    boxShadow: 24,
    p: 4,
};