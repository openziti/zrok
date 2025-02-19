import {AppBar, Box, Button, Grid2, Toolbar, Typography} from "@mui/material";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";
import zrokLogo from "./assets/zrok-1.0.0-rocket-green.svg";

interface NavBarProps {
    shareClick: () => void;
    accessClick: () => void;
}

const NavBar = ({ shareClick, accessClick }: NavBarProps) => {
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
                                <strong>z r o k &nbsp; Agent</strong>
                            </Grid2>
                        </Grid2>
                    </Typography>
                    <Grid2 container sx={{ flexGrow: 1 }}>
                        <Grid2 display="flex" justifyContent="right" size="grow">
                            <Button color="inherit" onClick={shareClick}><ShareIcon /></Button>
                        </Grid2>
                        <Grid2 display="flex" justifyContent="right">
                            <Button color="inherit" onClick={accessClick}><LanIcon /></Button>
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
        </Box>
    );
}

export default NavBar