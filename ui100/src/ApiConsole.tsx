import {useEffect, useRef, useState} from "react";
import {Configuration, MetadataApi} from "./api";
import {mergeVisualOverview, nodesEqual, VisualOverview} from "./model/visualizer.ts";
import {Box} from "@mui/material";
import NavBar from "./NavBar.tsx";
import {User} from "./model/user.ts";
import Visualizer from "./Visualizer.tsx";
import {Node} from "@xyflow/react";

interface ApiConsoleProps {
    user: User;
    logout: () => void;
}

const ApiConsole = ({ user, logout }: ApiConsoleProps) => {
    const [version, setVersion] = useState("no version set");
    const [overview, setOverview] = useState(new VisualOverview());
    const [selectedNode, setSelectedNode] = useState(null as Node);
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

    return (
        <div>
            <NavBar logout={logout} />
            <Box>
                <Visualizer vov={overview} onSelectionChanged={setSelectedNode} />
                <div>
                    <h1>Hello</h1>
                    <h2><pre>{JSON.stringify(selectedNode)}</pre></h2>
                </div>
            </Box>
        </div>
    );
}

export default ApiConsole;