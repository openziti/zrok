import {JSX, useCallback, useEffect, useRef, useState} from "react";
import {Graph, focusGraph, layout, mergeGraph, nodesEqual} from "./model/graph.ts";
import {Box, Button, Grid2, IconButton, Typography} from "@mui/material";
import OpenInFullIcon from "@mui/icons-material/OpenInFull";
import ErrorBoundary from "./ErrorBoundary.tsx";
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
    const selectedNodeRef = useRef<Node>(selectedNode);
    selectedNodeRef.current = selectedNode;
    const focusNodeId = useApiConsoleStore((state) => state.focusNodeId);
    const focusNodeIdRef = useRef<string | null>(focusNodeId);
    focusNodeIdRef.current = focusNodeId;
    const updateFocusNodeId = useApiConsoleStore((state) => state.updateFocusNodeId);
    const [mainPanel, setMainPanel] = useState(<Visualizer />);
    const [sidePanel, setSidePanel] = useState<JSX>(null);
    const [visualizerEnabled, setVisualizerEnabled] = useState<boolean>(true);
    const [panelMinimized, setPanelMinimized] = useState<boolean>(false);
    const panelMinimizedRef = useRef<boolean>(false);
    panelMinimizedRef.current = panelMinimized;
    const visualizerRef = useRef<boolean>(true);
    visualizerRef.current = visualizerEnabled;

    const applyFocusAndLayout = (graph: Graph, newFocusId: string | null) => {
        updateFocusNodeId(newFocusId);
        let graphToLayout = graph;
        if(newFocusId) {
            graphToLayout = focusGraph(graph, newFocusId);
        }
        let laidOut = layout(graphToLayout.nodes, graphToLayout.edges);
        let selected = laidOut.nodes.map((n) => ({
            ...n,
            selected: selectedNodeRef.current ? selectedNodeRef.current.id === n.id : false,
        }));
        updateNodes(selected);
        updateEdges(laidOut.edges);
    };

    const handleKeyPress = useCallback((event) => {
        if(event.ctrlKey === true && event.key === '`') {
            setVisualizerEnabled(!visualizerRef.current);
            return;
        }
        let tag = (event.target as HTMLElement)?.tagName?.toLowerCase();
        if(tag === "input" || tag === "textarea") return;
        if(event.key === 'f') {
            if(focusNodeIdRef.current) {
                applyFocusAndLayout(oldGraph.current, null);
            } else if(selectedNodeRef.current && selectedNodeRef.current.type !== "account") {
                applyFocusAndLayout(oldGraph.current, selectedNodeRef.current.id);
            }
            return;
        }
        if(event.key === 'p') {
            setPanelMinimized(!panelMinimizedRef.current);
            return;
        }
        if(event.key === 'Escape' && focusNodeIdRef.current) {
            applyFocusAndLayout(oldGraph.current, null);
            return;
        }
    }, []);

    const retrieveOverview = (signal?: AbortSignal) => {
        getMetadataApi(userRef.current).overview({ signal })
            .then(d => {
                updateLimited(d.accountLimited!);
                let newVov = mergeGraph(oldGraph.current, user, d.accountLimited!, d);
                if(!nodesEqual(oldGraph.current.nodes, newVov.nodes)) {
                    updateGraph(newVov);
                    oldGraph.current = newVov;

                    let graphToLayout = newVov;
                    if(focusNodeIdRef.current) {
                        if(!newVov.nodes.find(n => n.id === focusNodeIdRef.current)) {
                            updateFocusNodeId(null);
                        } else {
                            graphToLayout = focusGraph(newVov, focusNodeIdRef.current);
                        }
                    }

                    let laidOut = layout(graphToLayout.nodes, graphToLayout.edges);
                    let selected = laidOut.nodes.map((n) => ({
                        ...n,
                        selected: selectedNode ? selectedNode.id === n.id : false,
                    }));
                    updateNodes(selected);
                    updateEdges(laidOut.edges);
                }
            })
            .catch(() => {});
    }

    const retrieveSparklines = (signal?: AbortSignal) => {
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

        getMetadataApi(user).getSparklines({body: {environments: environments, shares: shares}}, { signal })
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
            .catch(() => {
            });
    }

    useEffect(() => {
        if(visualizerEnabled) {
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
        const controller = new AbortController();
        const doRetrieve = () => retrieveOverview(controller.signal);
        doRetrieve();
        let interval = setInterval(doRetrieve, 1000);
        return () => {
            controller.abort();
            clearInterval(interval);
        }
    }, []);

    useEffect(() => {
        const controller = new AbortController();
        let interval = setInterval(() => {
            retrieveSparklines(controller.signal);
        }, 5000);
        return () => {
            controller.abort();
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
            <div style={{ position: "relative" }}>
                <Grid2 container spacing={2} columns={{ xs: 4, sm: 10, md: 12 }}>
                    <Grid2 size="grow">
                        <ErrorBoundary fallback={({ reset }) => (
                            <Box sx={{ p: 3, textAlign: "center" }}>
                                <Typography color="error">The view encountered an error.</Typography>
                                <Button onClick={reset} variant="outlined" sx={{ mt: 1 }}>Try Again</Button>
                            </Box>
                        )}>
                            {mainPanel}
                        </ErrorBoundary>
                    </Grid2>
                    {sidePanel && !panelMinimized ? <Grid2 container size={4}><Grid2><ErrorBoundary key={selectedNode?.id}>{sidePanel}</ErrorBoundary></Grid2></Grid2> : null}
                </Grid2>
                {sidePanel && panelMinimized ? (
                    <div style={{ position: "absolute", top: 16, right: 16, zIndex: 5, display: "flex", alignItems: "center", gap: 4, background: "rgba(36, 23, 117, 0.85)", borderRadius: 8, padding: "4px 12px" }}>
                        <Typography variant="body2" sx={{ color: "#fff", whiteSpace: "nowrap" }}>
                            {selectedNode?.type}
                        </Typography>
                        <IconButton size="small" onClick={() => setPanelMinimized(false)} sx={{ color: "#fff", p: 0.25 }}>
                            <OpenInFullIcon sx={{ fontSize: 16 }} />
                        </IconButton>
                    </div>
                ) : null}
            </div>
        </div>
    );
}

export default ApiConsole;