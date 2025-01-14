import {AppBar, Box, Button, Grid2, Toolbar, Tooltip, Typography} from "@mui/material";
import LogoutIcon from "@mui/icons-material/Logout";
import VisualizerIcon from "@mui/icons-material/Hub";
import TabularIcon from "@mui/icons-material/TableRows";
import zroket from "./assets/zrok-1.0.0-rocket-green.svg";

interface NavBarProps {
    logout: () => void;
    visualizer: boolean;
    toggleMode: (boolean) => void;
}

const NavBar = ({ logout, visualizer, toggleMode }: NavBarProps) => {
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
                                <img src={zroket} height="30" />
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