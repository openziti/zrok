import * as metadata from './api/metadata';
import {useEffect, useLayoutEffect, useRef, useState} from "react";
import ReactFlow, {isNode, useNodesState} from "react-flow-renderer";
import dagre from 'dagre';

const Network = (props) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges] = useState([]);

    useEffect(() => {
        let mounted = true
        let interval = setInterval(() => {
            metadata.overview().then(resp => {
                let g = buildGraph(resp.data)
                setNodes(getLayout(g))
                setEdges(g.edges)
            })
        }, 1000)
        return () => {
            mounted = false
            clearInterval(interval)
        }
    }, [])

    return (
        <div className={"network"}>
            <h1>Network</h1>
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
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
            data: {label: 'Environment: ' + item.environment.zitiIdentityId},
            position: {x: (id * 25), y: 0},
            draggable: true
        });
        id++
        if(item.services != null) {
            item.services.forEach((item) => {
                if(item.active) {
                    out.nodes.push({
                        id: '' + id,
                        data: {label: 'Service: ' + item.zitiServiceId},
                        position: {x: (id * 25), y: 0},
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