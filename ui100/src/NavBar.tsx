import {AppBar, Box, Button, Grid2, IconButton, Toolbar, Typography} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import LogoutIcon from "@mui/icons-material/Logout";

interface NavBarProps {
    logout: () => void;
    version: string;
}

const NavBar = ({ logout, version }: NavBarProps) => {
    return (
        <Box ssx={{ flexGrow: 1 }}>
            <AppBar position="static">
                <Toolbar>
                    <IconButton size="large" edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }}>
                        <MenuIcon />
                    </IconButton>
                    <Typography variant="h6" sx={{ flexGrow: 1 }} display={{ xs: "none", sm: "none", md: "block" }}>
                        zrok { version !== "" ? " | " + version : ""}
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