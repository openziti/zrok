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
import useStore from "./model/store.ts";

interface ApiConsoleProps {
    user: User;
    logout: () => void;
}

const ApiConsole = ({ user, logout }: ApiConsoleProps) => {
    const overview = useStore((state) => state.overview);
    const updateOverview = useStore((state) => state.updateOverview);
    const oldVov = useRef<VisualOverview>(overview);
    const [selectedNode, setSelectedNode] = useState(null as Node);
    const [sidePanel, setSidePanel] = useState(<></>);
    const updateEnvironments = useStore((state) => state.updateEnvironments);

    const retrieveOverview = () => {
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
                    updateOverview(newVov);
                    oldVov.current = newVov;
                }
            })
            .catch(e => {
                console.log(e);
            });
    }

    useEffect(() => {
        retrieveOverview();
        let mounted = true;
        let interval = setInterval(() => {
            if(mounted) {
                retrieveOverview();
            }
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    const retrieveEnvironmentDetail = () => {
        let cfg = new Configuration({
            headers: {
                "X-TOKEN": user.token
            }
        });
        let metadata = new MetadataApi(cfg);
        metadata.getAccountDetail()
            .then(d => {
                updateEnvironments(d);
            })
            .catch(e => {
                console.log("environmentDetail", e);
            });
    }

    useEffect(() => {
        retrieveEnvironmentDetail();
        let interval = setInterval(() => {
            retrieveEnvironmentDetail();
        }, 5000);
        return () => {
            clearInterval(interval);
        }
    }, []);

    useEffect(() => {
        if(selectedNode) {
            switch(selectedNode.type) {
                case "account":
                    setSidePanel(<AccountPanel account={selectedNode} user={user} />);
                    break;

                case "environment":
                    setSidePanel(<EnvironmentPanel environment={selectedNode} user={user} />);
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