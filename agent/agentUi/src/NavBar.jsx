import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import {AppBar, Button, IconButton, Toolbar, Typography} from "@mui/material";
import MenuIcon from "@mui/icons-material/Menu";
import {Link} from "react-router-dom";
import ListIcon from "@mui/icons-material/List";
import LanIcon from "@mui/icons-material/Lan";
import ShareIcon from "@mui/icons-material/Share";

function NavBar() {
    const [version, setVersion] = useState("");

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
            if(mounted) {
                setVersion(data.v);
            }
        });
    }, []);

    return (
        <AppBar position={"static"}>
            <Toolbar>
                <IconButton
                    size={"large"}
                    edge={"start"}
                    color={"inherit"}
                    aria-label={"menu"}
                    sx={{mr: 2}}
                >
                    <MenuIcon/>
                </IconButton>
                <Typography variant={"p"} component={"div"} sx={{flexGrow: 1}}>
                    zrok Agent { version !== "" ? " | " + version : ""}
                </Typography>
                <Button color={"inherit"} component={Link} to={"/"}><ListIcon /></Button>
                <Button color={"inherit"}><LanIcon /></Button>
                <Button color={"inherit"}><ShareIcon /></Button>
            </Toolbar>
        </AppBar>
    )
}

export default NavBar;