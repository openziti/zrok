import "./App.css";
import "bootstrap/dist/css/bootstrap.min.css";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import DataTable from "react-data-table-component";
import ShareIcon from "@mui/icons-material/Share";
import MenuIcon from "@mui/icons-material/Menu";
import LanIcon from "@mui/icons-material/Lan";
import ListIcon from "@mui/icons-material/List";
import {AppBar, Button, IconButton, Toolbar, Typography} from "@mui/material";

function App() {
    const [version, setVersion] = useState("");
    const [shares, setShares] = useState([]);
    const [accesses, setAccesses] = useState([]);

    const shareColumns = [
        {
            name: 'Token',
            selector: row => row.reserved ? row.token+' (reserved)' : row.token
        },
        {
            name: 'Share Mode',
            selector: row => row.shareMode
        },
        {
            name: 'Backend Mode',
            selector: row => row.backendMode
        },
        {
            name: 'Target',
            selector: row => row.backendEndpoint,
        },
        {
            name: 'Frontend Endpoints',
            selector: row => <div>{row.shareMode === "public" ? row.frontendEndpoint.map((fe) => <a href={fe.toString()} target={"_"}>{fe}</a>) : "---"}</div>,
            grow: 2
        }
    ];

    const accessColumns = [
        {
            name: 'Frontend Token',
            selector: row => row.frontendToken
        },
        {
            name: 'Token',
            selector: row => row.token
        },
        {
            name: 'Bind Address',
            selector: row => row.bindAddress
        },
    ];

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
           if(mounted) {
               setVersion(data.v);
           }
        });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                if(mounted) {
                    setShares(data.shares);
                    setAccesses(data.accesses);
                }
            });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    });

    return (
        <>
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
                    <Button color={"inherit"}><ListIcon /></Button>
                    <Button color={"inherit"}><LanIcon /></Button>
                    <Button color={"inherit"}><ShareIcon /></Button>
                </Toolbar>
            </AppBar>

            <div class={"info"}>
                <h2>Shares</h2>
                <DataTable
                    columns={shareColumns}
                    data={shares}
                    noDataComponent={<div/>}
                />
            </div>

            <div class={"info"}>
                <h2>Accesses</h2>
                <DataTable
                    columns={accessColumns}
                    data={accesses}
                    noDataComponent={<div/>}
                />
            </div>

            <code>Connected Version: {version}</code>
        </>
    )
}

export default App;
