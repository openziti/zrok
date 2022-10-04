import Environments from './Environments';
import * as metadata from './api/metadata';
import {useEffect, useLayoutEffect, useRef, useState} from "react";
import ReactFlow, {isNode, useNodesState, useReactFlow} from "react-flow-renderer";
import dagre from 'dagre';
import { mdiDesktopClassic, mdiAccessPointNetwork } from '@mdi/js';
import Icon from "@mdi/react";

const Network = () => {
    const [overview, setOverview] = useState([]);
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges] = useState([]);
    const reactFlow = useReactFlow();

    useEffect(() => {
        let mounted = true
        metadata.overview().then(resp => {
            if(mounted) {
                setOverview(resp.data)

                let g = buildGraph(resp.data)
                setNodes(getLayout(g))
                setEdges(g.edges)
                reactFlow.fitView({maxZoom: 1})
            }
        });
    })

    useEffect(() => {
        let mounted = true
        let interval = setInterval(() => {
            metadata.overview().then(resp => {
                if(mounted) {
                    setOverview(resp.data)

                    let g = buildGraph(resp.data)
                    setNodes(getLayout(g))
                    setEdges(g.edges)
                }
            })
        }, 1000)
        return () => {
            mounted = false
            clearInterval(interval)
        }
    }, [])

    return (
        <div>
            <div className={"network"}>
                <h1>Network</h1>
                <ReactFlow
                    nodes={nodes}
                    edges={edges}
                    onNodesChange={onNodesChange}
                />
            </div>
            <Environments
                overview={overview}
            />
        </div>
    )
}

function buildGraph(overview) {
    let out = {
        nodes: [],
        edges: []
    }
    let id = 1
    overview.forEach((item) => {
        let envId = id
        out.nodes.push({
            id: '' + envId,
            data: { label: <div><Icon path={mdiDesktopClassic} size={0.75} className={"flowNode"}/> { item.environment.description } </div> },
            position: { x: (id * 25), y: 0 },
            style: { width: 'fit-content', backgroundColor: '#aaa', color: 'white' },
            type: 'input',
            draggable: true
        });
        id++
        if(item.services != null) {
            item.services.forEach((item) => {
                if(item.active) {
                    out.nodes.push({
                        id: '' + id,
                        data: {label: <div><Icon path={mdiAccessPointNetwork} size={0.75} className={"flowNode"}/> { item.frontend }</div>},
                        position: {x: (id * 25), y: 0},
                        style: { width: 'fit-content', backgroundColor: '#9367ef', color: 'white' },
                        type: 'output',
                        draggable: true
                    })
                    out.edges.push({
                        id: 'e' + envId + '-' + id,
                        source: '' + envId,
                        target: '' + id,
                        animated: true
                    })
                    id++
                }
            });
        }
    });
    return out
}

const nodeWidth = 215;
const nodeHeight = 75;

function getLayout(overview) {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setGraph({ rankdir: 'TB' });
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    overview.nodes.forEach((n) => {
        dagreGraph.setNode(n.id, { width: nodeWidth, height: nodeHeight });
    })
    overview.edges.forEach((e) => {
        dagreGraph.setEdge(e.source, e.target);
    })
    dagre.layout(dagreGraph);

    return overview.nodes.map((n) => {
        const nodeWithPosition = dagreGraph.node(n.id);
        n.targetPosition = 'top';
        n.sourcePosition = 'bottom';
        n.position = {
            x: nodeWithPosition.x - (nodeWidth / 2) + (Math.random() / 1000) + 50,
            y: nodeWithPosition.y - (nodeHeight / 2) + 50,
        }
        return n;
    });
}


export default Network;