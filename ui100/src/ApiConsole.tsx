import {useEffect, useRef, useState} from "react";
import {Configuration, MetadataApi} from "./api";
import {mergeVisualOverview, nodesEqual, VisualOverview} from "./model/visualizer.ts";
import {Grid2} from "@mui/material";
import NavBar from "./NavBar.tsx";
import {User} from "./model/user.ts";
import Visualizer from "./Visualizer.tsx";
import {Node} from "@xyflow/react";
import AccountPanel from "./AccountPanel.tsx";
import EnvironmentPanel from "./EnvironmentPanel.tsx";
import SharePanel from "./SharePanel.tsx";
import AccessPanel from "./AccessPanel.tsx";

interface ApiConsoleProps {
    user: User;
    logout: () => void;
}

const ApiConsole = ({ user, logout }: ApiConsoleProps) => {
    const [version, setVersion] = useState("no version set");
    const [overview, setOverview] = useState(new VisualOverview());
    const [selectedNode, setSelectedNode] = useState(null as Node);
    const [sidePanel, setSidePanel] = useState(<></>);
    const oldVov = useRef<VisualOverview>(overview);

    useEffect(() => {
        let api = new MetadataApi();
        api.version()
            .then(d => {
                setVersion(d);
            })
            .catch(e => {
                console.log(e);
            });
    }, []);

    const update = () => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let api = new MetadataApi(cfg);
        api.overview()
            .then(d => {
                let newVov = mergeVisualOverview(oldVov.current, user, d.accountLimited!, d);
                if(!nodesEqual(oldVov.current.nodes, newVov.nodes)) {
                    console.log("refreshed vov", oldVov.current.nodes, newVov.nodes);
                    setOverview(newVov);
                    oldVov.current = newVov;
                }
            })
            .catch(e => {
                console.log(e);
            });
    }

    useEffect(() => {
        update();
        let mounted = true;
        let interval = setInterval(() => {
            if(mounted) {
                update();
            }
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    useEffect(() => {
        if(selectedNode) {
            switch(selectedNode.type) {
                case "account":
                    setSidePanel(<AccountPanel account={selectedNode} />);
                    break;

                case "environment":
                    setSidePanel(<EnvironmentPanel environment={selectedNode} />);
                    break;

                case "share":
                    setSidePanel(<SharePanel share={selectedNode} />);
                    break;

                case "access":
                    setSidePanel(<AccessPanel access={selectedNode} />);
                    break;
            }
        } else {
            setSidePanel(<></>);
        }
    }, [selectedNode]);

    return (
        <div>
            <NavBar logout={logout} />
            <Grid2 container spacing={2} columns={{ xs: 4, sm: 10, md: 12 }}>
                <Grid2 size="grow">
                    <Visualizer vov={overview} onSelectionChanged={setSelectedNode} />
                </Grid2>
                <Grid2 size={4}>
                    {sidePanel}
                </Grid2>
            </Grid2>
        </div>
    );
}

export default ApiConsole;