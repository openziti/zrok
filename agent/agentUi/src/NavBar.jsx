import {AppBar, Button, IconButton, Toolbar, Typography} from "@mui/material";
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
                <Typography variant="p" component={"div"} sx={{flexGrow: 1}}>
                    zrok Agent { props.version !== "" ? " | " + props.version : ""}
                </Typography>
                <Button color="inherit" onClick={props.shareClick}><ShareIcon /></Button>
                <Button color="inherit" onClick={props.accessClick}><LanIcon /></Button>
            </Toolbar>
        </AppBar>
    )
}

export default NavBar;