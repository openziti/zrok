import {JSX, useCallback, useEffect, useRef, useState} from "react";
import {Graph, layout, mergeGraph, nodesEqual} from "./model/graph.ts";
import {Grid2} from "@mui/material";
import NavBar from "./NavBar.tsx";
import Visualizer from "./Visualizer.tsx";
import AccountPanel from "./AccountPanel.tsx";
import EnvironmentPanel from "./EnvironmentPanel.tsx";
import SharePanel from "./SharePanel.tsx";
import AccessPanel from "./AccessPanel.tsx";
import useApiConsoleStore from "./model/store.ts";
import TabularView from "./TabularView.tsx";
import {Node} from "@xyflow/react";
import {getMetadataApi} from "./model/api.ts";
import {User} from "./model/user.ts";

interface ApiConsoleProps {
    logout: () => void;
}

const ApiConsole = ({ logout }: ApiConsoleProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const userRef = useRef<User>(user);
    const updateLimited = useApiConsoleStore((state) => state.updateLimited);
    const graph = useApiConsoleStore((state) => state.graph);
    const updateGraph = useApiConsoleStore((state) => state.updateGraph);
    const oldGraph = useRef<Graph>(graph);
    const sparkdata = useApiConsoleStore((state) => state.sparkdata);
    const sparkdataRef = useRef<Map<string, Number[]>>();
    sparkdataRef.current = sparkdata;
    const updateSparkdata = useApiConsoleStore((state) => state.updateSparkdata);
    const nodes = useApiConsoleStore((state) => state.nodes);
    const nodesRef = useRef<Node[]>();
    nodesRef.current = nodes;
    const updateNodes = useApiConsoleStore((state) => state.updateNodes);
    const updateEdges = useApiConsoleStore((state) => state.updateEdges);
    const selectedNode = useApiConsoleStore((state) => state.selectedNode);
    const [mainPanel, setMainPanel] = useState(<Visualizer />);
    const [sidePanel, setSidePanel] = useState<JSX>(null);
    const [visualizerEnabled, setVisualizerEnabled] = useState<boolean>(true);

    let visualizer = true;
    const handleKeyPress = useCallback((event) => {
        if(event.ctrlKey === true && event.key === '`') {
            setVisualizerEnabled(!visualizer);
            visualizer = !visualizer;
            if(visualizer) {
                setMainPanel(<Visualizer />);
            } else {
                setMainPanel(<TabularView />);
            }
        }
    }, []);

    const retrieveOverview = () => {
        getMetadataApi(userRef.current).overview()
            .then(d => {
                updateLimited(d.accountLimited!);
                let newVov = mergeGraph(oldGraph.current, user, d.accountLimited!, d);
                if(!nodesEqual(oldGraph.current.nodes, newVov.nodes)) {
                    console.log("refreshed vov", oldGraph.current.nodes, newVov.nodes);
                    updateGraph(newVov);
                    oldGraph.current = newVov;

                    let laidOut = layout(newVov.nodes, newVov.edges);
                    let selected = laidOut.nodes.map((n) => ({
                        ...n,
                        selected: selectedNode ? selectedNode.id === n.id : false,
                    }));
                    updateNodes(selected);
                    updateEdges(laidOut.edges);
                }
            })
            .catch(e => {
                console.log(e);
            });
    }

    const retrieveSparklines = () => {
        let environments: string[] = [];
        let shares: string[] = [];
        if(nodesRef.current) {
            nodesRef.current.map(node => {
                if(node.type === "environment") {
                    environments.push(node.id);
                }
                if(node.type === "share") {
                    shares.push(node.id);
                }
            });
        }

        getMetadataApi(user).getSparklines({body: {environments: environments, shares: shares}})
            .then(d => {
                if(d.sparklines) {
                    let sparkdataIn = new Map<string, Number[]>();
                    d.sparklines!.forEach(s => {
                        let activity = new Array<Number>(31);
                        if(s.samples) {
                            s.samples?.forEach((sample, i) => {
                                let v = 0;
                                v += sample.rx! ? sample.rx! : 0;
                                v += sample.tx! ? sample.tx! : 0;
                                activity[i] = v;
                            });
                            sparkdataIn.set(s.id!, activity);
                        }
                    });
                    updateSparkdata(sparkdataIn);
                } else {
                    updateSparkdata(new Map<string, Number[]>());
                }
            })
            .catch(e => {
                console.log("getSparklines", e);
            });
    }

    useEffect(() => {
        visualizer = visualizerEnabled;
        if(visualizer) {
            setMainPanel(<Visualizer />);
        } else {
            setMainPanel(<TabularView />);
        }
    }, [visualizerEnabled]);

    useEffect(() => {
        document.addEventListener('keydown', handleKeyPress);
        return () => {
            document.removeEventListener('keydown', handleKeyPress);
        };
    }, [handleKeyPress]);

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

    useEffect(() => {
        let interval = setInterval(() => {
            retrieveSparklines();
        }, 5000);
        return () => {
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
            setSidePanel(null);
        }
    }, [selectedNode]);

    return (
        <div>
            <NavBar logout={logout} visualizer={visualizerEnabled} toggleMode={setVisualizerEnabled} />
            <Grid2 container spacing={2} columns={{ xs: 4, sm: 10, md: 12 }}>
                <Grid2 size="grow">
                    {mainPanel}
                </Grid2>
                {sidePanel ? <Grid2 container size={5}><Grid2 >{sidePanel}</Grid2></Grid2> : null}
            </Grid2>
        </div>
    );
}

export default ApiConsole;