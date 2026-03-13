import {useCallback, useEffect, useRef, useState} from "react";
import {Graph, focusGraph, layout, mergeGraph, nodesEqual} from "./model/graph.ts";
import {Box, Button, IconButton, Typography} from "@mui/material";
import {alpha} from "@mui/material/styles";
import {COLORS} from "./styling/theme.ts";
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
import {isAbortError} from "./model/errors.ts";

interface ApiConsoleProps {
    logout: () => void;
}

const ApiConsole = ({ logout }: ApiConsoleProps) => {
    const user = useApiConsoleStore((state) => state.user);
    const userRef = useRef<User | null>(user);
    const updateLimited = useApiConsoleStore((state) => state.updateLimited);
    const updateEnvironments = useApiConsoleStore((state) => state.updateEnvironments);
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
    const selectedNodeRef = useRef<Node | null>(selectedNode);
    selectedNodeRef.current = selectedNode;
    const focusNodeId = useApiConsoleStore((state) => state.focusNodeId);
    const focusNodeIdRef = useRef<string | null>(focusNodeId);
    focusNodeIdRef.current = focusNodeId;
    const updateFocusNodeId = useApiConsoleStore((state) => state.updateFocusNodeId);
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
        if (!userRef.current) return Promise.resolve();
        const metadataApi = getMetadataApi(userRef.current);
        return Promise.all([
            metadataApi.overview({ signal }),
            metadataApi.getAccountDetail({ signal }),
        ])
            .then(([d, accountDetail]) => {
                updateEnvironments(accountDetail);
                return d;
            })
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
            });
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

        return getMetadataApi(user!).getSparklines({body: {environments: environments, shares: shares}}, { signal })
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
            });
    }

    const renderSidePanel = () => {
        if (!selectedNode) return null;
        switch (selectedNode.type) {
            case "account": return <AccountPanel account={selectedNode} />;
            case "environment": return <EnvironmentPanel environment={selectedNode} />;
            case "share": return <SharePanel share={selectedNode} />;
            case "access": return <AccessPanel access={selectedNode} />;
            default: return null;
        }
    };

    useEffect(() => {
        document.addEventListener('keydown', handleKeyPress);
        return () => {
            document.removeEventListener('keydown', handleKeyPress);
        };
    }, [handleKeyPress]);

    useEffect(() => {
        const controller = new AbortController();
        let overviewTimeout: ReturnType<typeof setTimeout>;
        let sparkTimeout: ReturnType<typeof setTimeout>;
        let overviewDelay = 5000;
        let sparkDelay = 15000;
        let disposed = false;

        const pollOverview = () => {
            retrieveOverview(controller.signal)
                .then(() => { overviewDelay = 5000; })
                .catch((e) => {
                    if (isAbortError(e)) {
                        return;
                    }
                    overviewDelay = Math.min(overviewDelay * 2, 30000);
                })
                .finally(() => {
                    if (!disposed && !controller.signal.aborted) {
                        overviewTimeout = setTimeout(pollOverview, overviewDelay);
                    }
                });
        };

        const pollSparklines = () => {
            retrieveSparklines(controller.signal)
                .then(() => { sparkDelay = 15000; })
                .catch((e) => {
                    if (isAbortError(e)) {
                        return;
                    }
                    sparkDelay = Math.min(sparkDelay * 2, 30000);
                })
                .finally(() => {
                    if (!disposed && !controller.signal.aborted) {
                        sparkTimeout = setTimeout(pollSparklines, sparkDelay);
                    }
                });
        };

        // initial load: overview first, then sparklines once nodes are populated
        retrieveOverview(controller.signal)
            .then(() => { overviewDelay = 5000; })
            .catch((e) => {
                if (isAbortError(e)) {
                    return;
                }
                overviewDelay = Math.min(overviewDelay * 2, 30000);
            })
            .then(() => retrieveSparklines(controller.signal)
                .then(() => { sparkDelay = 15000; })
                .catch((e) => {
                    if (isAbortError(e)) {
                        return;
                    }
                    sparkDelay = Math.min(sparkDelay * 2, 30000);
                }))
            .finally(() => {
                if (!disposed && !controller.signal.aborted) {
                    overviewTimeout = setTimeout(pollOverview, overviewDelay);
                    sparkTimeout = setTimeout(pollSparklines, sparkDelay);
                }
            });

        return () => {
            disposed = true;
            controller.abort();
            clearTimeout(overviewTimeout);
            clearTimeout(sparkTimeout);
        };
    }, []);

    return (
        <Box
            sx={{
                height: "100dvh",
                display: "flex",
                flexDirection: "column",
                overflow: "hidden",
                boxSizing: "border-box",
                py: "15px",
                px: "15px",
                gap: "15px",
            }}
        >
            <NavBar logout={logout} visualizer={visualizerEnabled} toggleMode={setVisualizerEnabled} />
            <Box sx={{ position: "relative", flex: 1, minHeight: 0, overflow: "hidden" }}>
                <Box
                    sx={{
                        display: "grid",
                        gridTemplateColumns: selectedNode && !panelMinimized ? "minmax(0, 1fr) 360px" : "minmax(0, 1fr)",
                        gap: "15px",
                        height: "100%",
                        minHeight: 0,
                    }}
                >
                    <Box sx={{ minWidth: 0, minHeight: 0, overflow: "hidden" }}>
                        <ErrorBoundary fallback={({ reset }) => (
                            <Box sx={{ p: 3, textAlign: "center" }}>
                                <Typography color="error">The view encountered an error.</Typography>
                                <Button onClick={reset} variant="outlined" sx={{ mt: 1 }}>Try Again</Button>
                            </Box>
                        )}>
                            {visualizerEnabled ? <Visualizer /> : <TabularView />}
                        </ErrorBoundary>
                    </Box>
                    {selectedNode && !panelMinimized ? (
                        <Box
                            sx={{
                                minHeight: 0,
                                minWidth: 0,
                                overflow: "auto",
                            }}
                        >
                            <ErrorBoundary key={selectedNode?.id}>{renderSidePanel()}</ErrorBoundary>
                        </Box>
                    ) : null}
                </Box>
                {selectedNode && panelMinimized ? (
                    <Box sx={{ position: "absolute", top: 16, right: 16, zIndex: 5, display: "flex", alignItems: "center", gap: 4, background: alpha(COLORS.primary, 0.85), borderRadius: 8, padding: "4px 12px" }}>
                        <Typography variant="body2" sx={{ color: 'common.white', whiteSpace: "nowrap" }}>
                            {selectedNode?.type}
                        </Typography>
                        <IconButton size="small" aria-label="Expand panel" onClick={() => setPanelMinimized(false)} sx={{ color: 'common.white', p: 0.25 }}>
                            <OpenInFullIcon sx={{ fontSize: 16 }} />
                        </IconButton>
                    </Box>
                ) : null}
            </Box>
        </Box>
    );
}

export default ApiConsole;
