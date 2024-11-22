import {useEffect, useState} from "react";
import {Configuration, MetadataApi} from "./api";
import buildVisualizerGraph from "./model/visualizer.ts";
import {GraphCanvas} from "reagraph";
import {Box} from "@mui/material";
import NavBar from "./NavBar.tsx";

const ApiConsole = () => {
    const [version, setVersion] = useState("no version set");
    const [nodes, setNodes] = useState([]);
    const [edges, setEdges] = useState([]);

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
                    // ignorable token, local development environment
                    "X-TOKEN": "q9bwDQqMQ6K6"
                }
            });
            let api = new MetadataApi(cfg);
            api.overview()
                .then(d => {
                    console.log(d);
                    let graph = buildVisualizerGraph(d);

                    setNodes(graph.nodes);
                    setEdges(graph.edges);
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
            <NavBar version={version} />
            <Box>
                <div style={{position: "relative", width: "100%", height: "500px"}}>
                    <GraphCanvas nodes={nodes} edges={edges}/>
                </div>
            </Box>
        </div>
    );
}

export default ApiConsole;