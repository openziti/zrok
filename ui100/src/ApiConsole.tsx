import {useEffect, useState} from "react";
import {Configuration, MetadataApi} from "./api";
import buildVisualizerGraph, {VisualOverview} from "./model/visualizer.ts";
import {Box} from "@mui/material";
import NavBar from "./NavBar.tsx";
import {User} from "./model/user.ts";
import Visualizer from "./Visualizer.tsx";

interface ApiConsoleProps {
    user: User;
    logout: () => void;
}

const ApiConsole = ({ user, logout }: ApiConsoleProps) => {
    const [version, setVersion] = useState("no version set");
    const [overview, setOverview] = useState(new VisualOverview());

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

    useEffect(() => {
        let interval = setInterval(() => {
            let cfg = new Configuration({
                headers: {
                    "X-TOKEN": user.token
                }
            });
            let api = new MetadataApi(cfg);
            api.overview()
                .then(d => {
                    setOverview(buildVisualizerGraph(d));
                })
                .catch(e => {
                    console.log(e);
                });
        }, 1000);
        return () => {
            clearInterval(interval);
        }
    }, []);

    return (
        <div>
            <NavBar logout={logout} />
            <Box>
                <Visualizer overview={overview} />
            </Box>
        </div>
    );
}

export default ApiConsole;