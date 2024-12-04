import {AppBar, Box, Button, Grid2, IconButton, Toolbar, Typography} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import LogoutIcon from "@mui/icons-material/Logout";
import zroket from "./assets/zrok-1.0.0-rocket-white.svg";

interface NavBarProps {
    logout: () => void;
}

const NavBar = ({ logout }: NavBarProps) => {
    return (
        <Box ssx={{ flexGrow: 1 }}>
            <AppBar position="static">
                <Toolbar>
                    <IconButton size="large" edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }}>
                        <MenuIcon />
                    </IconButton>
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
                            <Button color="inherit" onClick={logout}><LogoutIcon /></Button>
                        </Grid2>
                    </Grid2>
                </Toolbar>
            </AppBar>
        </Box>
    );
}

export default NavBar;