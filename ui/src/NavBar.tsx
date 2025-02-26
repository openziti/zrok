import {AppBar, Box, Button, Grid2, Toolbar, Tooltip, Typography} from "@mui/material";
import LogoutIcon from "@mui/icons-material/Logout";
import VisualizerIcon from "@mui/icons-material/Hub";
import TabularIcon from "@mui/icons-material/TableRows";
import LimitIcon from "@mui/icons-material/CrisisAlert";
import HelpIcon from "@mui/icons-material/LiveHelp";
import zrokLogo from "./assets/zrok-1.0.0-rocket-green.svg";
import useApiConsoleStore from "./model/store.ts";
import BandwidthLimitedModal from "./BandwidthLimitedModal.tsx";
import {useEffect, useState} from "react";
import GettingStartedModal from "./GettingStartedModal.tsx";

interface NavBarProps {
    logout: () => void;
    visualizer: boolean;
    toggleMode: (boolean) => void;
}

const NavBar = ({ logout, visualizer, toggleMode }: NavBarProps) => {
    const nodes = useApiConsoleStore((state) => state.nodes);
    const limited = useApiConsoleStore((state) => state.limited);
    const [limitedModalOpen, setLimitedModalOpen] = useState<boolean>(false);
    const openLimitedModal = () => {
        setLimitedModalOpen(true);
    }
    const closeLimitedModal = () => {
        setLimitedModalOpen(false);
    }
    const [gettingStartedOpen, setGettingStartedOpen] = useState<boolean>(false);
    const openGettingStarted = () => {
        setGettingStartedOpen(true);
    }
    const closeGettingStarted = () => {
        setGettingStartedOpen(false);
    }

    useEffect(() => {
        if(!limited) {
            closeLimitedModal();
        }
    }, [limited])

    const limitedIndicator = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Bandwidth Limit Reached!">
                <Button color="error" onClick={openLimitedModal}><LimitIcon /></Button>
            </Tooltip>
        </Grid2>
    );

    const gettingStartedButton = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Getting Started Wizard">
                <Button style={{ backgroundColor: "#9bf316", color: "black" }} onClick={openGettingStarted}>CLICK HERE TO GET STARTED!</Button>
            </Tooltip>
        </Grid2>
    );

    const helpButton = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Getting Started Wizard">
                <Button style={{ color: "#9bf316" }} onClick={openGettingStarted}><HelpIcon /></Button>
            </Tooltip>
        </Grid2>
    );

    const handleClick = () => {
        toggleMode(!visualizer);
    }

    return (
        <>
            <Box sx={{ flexGrow: 1 }}>
                <AppBar position="static">
                    <Toolbar>
                        <Typography variant="h6" sx={{ flexGrow: 1 }}>
                            <Grid2 container sx={{ flexGrow: 1 }}>
                                <Grid2 display="flex" justifyContent="left">
                                    <img src={zrokLogo} height="30" />
                                </Grid2>
                                <Grid2 display="flex" justifyContent="left" size="grow" sx={{ ml: 3 }} color="#9bf316">
                                    <strong>z r o k</strong>
                                </Grid2>
                            </Grid2>
                        </Typography>
                        <Grid2 container sx={{ flexGrow: 1 }}>
                            <Grid2 display="flex" justifyContent="right" size="grow">
                                <Tooltip title="Toggle Interface Mode (Ctrl-`)">
                                    <Button color="inherit" onClick={handleClick}>{ visualizer ? <VisualizerIcon /> : <TabularIcon /> }</Button>
                                </Tooltip>
                            </Grid2>
                            { limited ? limitedIndicator : null }
                            { !nodes || nodes.length > 1 ? helpButton : gettingStartedButton }
                            <Grid2 display="flex" justifyContent="right">
                                <Tooltip title="Logout">
                                    <Button color="inherit" onClick={logout}><LogoutIcon /></Button>
                                </Tooltip>
                            </Grid2>
                        </Grid2>
                    </Toolbar>
                </AppBar>
            </Box>
            <BandwidthLimitedModal close={closeLimitedModal} isOpen={limitedModalOpen} />
            <GettingStartedModal close={closeGettingStarted} isOpen={gettingStartedOpen} />
        </>
    );
}

export default NavBar;