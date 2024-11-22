import {AppBar, Box, IconButton, Toolbar, Typography} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";

interface NavBarProps {
    version: string;
}

const NavBar = ({ version }: NavBarProps) => {
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
                </Toolbar>
            </AppBar>
        </Box>
    );
}

export default NavBar;