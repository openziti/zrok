import {AppBar, Box, Button, Grid2, IconButton, Toolbar, Typography} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";

const NavBar = (props) => {
    return (
        <AppBar position={"static"}>
            <Toolbar>
                <IconButton
                    size="large"
                    edge={"start"}
                    color="inherit"
                    aria-label={"menu"}
                    sx={{mr: 2}}
                >
                    <MenuIcon/>
                </IconButton>
                <Typography variant="p" component={"div"} sx={{ flexGrow: 1 }} display={{ xs: "none", sm: "block" }}>
                    zrok Agent { props.version !== "" ? " | " + props.version : ""}
                </Typography>
                <Grid2 container sx={{ flexGrow: 1 }}>
                    <Grid2 display="flex" justifyContent="right" size="grow">
                        <Button color="inherit" onClick={props.shareClick}><ShareIcon /></Button>
                    </Grid2>
                    <Grid2 display="flex" justifyContent="right">
                        <Button color="inherit" onClick={props.accessClick}><LanIcon /></Button>
                    </Grid2>
                </Grid2>
            </Toolbar>
        </AppBar>
    )
}

export default NavBar;