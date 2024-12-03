import {createTheme} from "@mui/material";
import {Theme} from "reagraph";

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

export const reagraphTheme: Theme = {
    canvas: {
        background: '#fff',
        fog: '#fff'
    },
    node: {
        fill: '#241775',
        activeFill: '#9bf316',
        opacity: 1,
        selectedOpacity: 1,
        inactiveOpacity: 0.2,
        label: {
            color: '#241775',
            stroke: '#fff',
            activeColor: '#9bf316'
        },
        subLabel: {
            color: '#241775',
            stroke: '#eee',
            activeColor: '#9bf316'
        }
    },
    lasso: {
        border: '1px solid #55aaff',
        background: 'rgba(75, 160, 255, 0.1)'
    },
    ring: {
        fill: '#D8E6EA',
        activeFill: '#1DE9AC'
    },
    edge: {
        fill: '#D8E6EA',
        activeFill: '#1DE9AC',
        opacity: 1,
        selectedOpacity: 1,
        inactiveOpacity: 0.1,
        label: {
            stroke: '#fff',
            color: '#2A6475',
            activeColor: '#1DE9AC'
        }
    },
    arrow: {
        fill: '#D8E6EA',
        activeFill: '#1DE9AC'
    },
    cluster: {
        stroke: '#D8E6EA',
        opacity: 1,
        selectedOpacity: 1,
        inactiveOpacity: 0.1,
        label: {
            stroke: '#fff',
            color: '#2A6475'
        }
    }
};