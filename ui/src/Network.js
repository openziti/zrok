import ReactFlow, {useNodesState} from "react-flow-renderer";
import * as metadata from './api/metadata';
import {useEffect, useState} from "react";

const Network = (props) => {
    const [nodes, setNodes, onNodesChange] = useNodesState([])
    const [edges, setEdges] = useState([]);

    useEffect(() => {
        let mounted = true
        metadata.overview().then(resp => {
            let ovr = buildGraph(resp.data)
            setNodes(ovr.nodes)
            setEdges(ovr.edges)
            console.log('nodes', ovr.nodes);
        })
        return () => {
            mounted = false
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
        item.services.forEach((item) => {
            out.nodes.push({
                id: '' + id,
                data: {label: 'Service: ' + item.zitiServiceId},
                position: {x: (id * 25), y: 0},
                draggable: true
            })
            out.edges.push({
                id: 'e' + envId + '-' + id,
                source: '' + id,
                target: '' + envId,
                animated: true
            })
            id++
        });
    });
    return out
}

export default Network;