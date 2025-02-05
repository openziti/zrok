import {AppBar, Box, Button, Grid2, Toolbar, Tooltip, Typography} from "@mui/material";
import LogoutIcon from "@mui/icons-material/Logout";
import VisualizerIcon from "@mui/icons-material/Hub";
import TabularIcon from "@mui/icons-material/TableRows";
import LimitIcon from "@mui/icons-material/CrisisAlert";
import zrokLogo from "./assets/zrok-1.0.0-rocket-green.svg";
import useApiConsoleStore from "./model/store.ts";

interface NavBarProps {
    logout: () => void;
    visualizer: boolean;
    toggleMode: (boolean) => void;
}

const NavBar = ({ logout, visualizer, toggleMode }: NavBarProps) => {
    const limited = useApiConsoleStore((state) => state.limited);

    const limitedIndicator = (
        <Grid2 display="flex" justifyContent="right">
            <Tooltip title="Bandwidth Limit Reached!">
                <Button color="error" ><LimitIcon /></Button>
            </Tooltip>
        </Grid2>
    );

    const handleClick = () => {
        toggleMode(!visualizer);
    }

    return (
        <Box ssx={{ flexGrow: 1 }}>
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
                            <Button variant="outline" color="inherit">CLICK HERE TO GET STARTED!</Button>
                        </Grid2>
                        { limited ? limitedIndicator : null }
                        <Grid2 display="flex" justifyContent="right">
                            <Tooltip title="Toggle Interface Mode (Ctrl-`)">
                                <Button color="inherit" onClick={handleClick}>{ visualizer ? <VisualizerIcon /> : <TabularIcon /> }</Button>
                            </Tooltip>
                        </Grid2>
                        <Grid2 display="flex" justifyContent="right">
                            <Tooltip title="Logout">
                                <Button color="inherit" onClick={logout}><LogoutIcon /></Button>
                            </Tooltip>
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
        </Box>
    );
}

export default NavBar;