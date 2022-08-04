import * as metadata from './api/metadata';
import {useEffect, useLayoutEffect, useRef, useState} from "react";
import ForceGraph2D from 'react-force-graph-2d';

let g1 = {}

const Network = (props) => {
    const ref = useRef();
    const [graph, setGraph] = useState({nodes: [], links: []})

    useEffect(() => {
        let mounted = true
        let g1 = graph
        let interval = setInterval(() => {
            metadata.overview().then(resp => {
                let g = buildGraph(resp.data)
                if(!compareGraphs(g, g1)) {
                    setGraph(g)
                    g1 = g
                }
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
            <ForceGraph2D
                ref={ref}
                width={1024}
                height={300}
                graphData={graph}
                nodeDefaultSize={[100, 50]}
                nodeCanvasObject={(node, ctx, globalScale) => {
                    const label = node.name;
                    const fontSize = 12/globalScale;
                    ctx.font = `${fontSize}px JetBrains Mono`;
                    const textWidth = ctx.measureText(label).width;
                    const bckgDimensions = [textWidth, fontSize].map(n => n + fontSize * 2.2); // some padding

                    ctx.fillStyle = '#3b2693';
                    ctx.strokeStyle = '#3b2693'
                    ctx.lineWidth = 0.5
                    ctx.fillRect(node.x - bckgDimensions[0] / 2, node.y - bckgDimensions[1] / 2, ...bckgDimensions);

                    ctx.textAlign = 'center';
                    ctx.textBaseline = 'middle';
                    ctx.fillStyle = 'white';
                    ctx.fillText(label, node.x, node.y);

                    node.__bckgDimensions = bckgDimensions; // to re-use in nodePointerAreaPaint
                }}
                onEngineStop={() => {
                    ref.current.zoomToFit(200);
                }}
            />
        </div>
    )
}

function buildGraph(overview) {
    let out = {
        nodes: [],
        links: []
    }
    overview.forEach((item) => {
        out.nodes.push({
            id: item.environment.zitiIdentityId,
            name: 'Environment: ' + item.environment.zitiIdentityId
        });
        if(item.services != null) {
            item.services.forEach((svc) => {
                if(svc.active) {
                    out.nodes.push({
                        id: svc.zitiServiceId,
                        name: 'Service: ' + svc.zitiServiceId
                    })
                    out.links.push({
                        source: item.environment.zitiIdentityId,
                        target: svc.zitiServiceId
                    })
                }
            });
        }
    });
    return out
}

function compareGraphs(g, g1) {
    if(g.nodes.length !== g1.nodes.length) return false;
    for(let i = 0; i < g.nodes.length; i++) {
        if(!compareNodes(g.nodes[i], g1.nodes[i])) return false;
    }
    return true
}

function compareNodes(n, n1) {
    if(n.id !== n1.id) return false;
    if(n.name !== n1.name) return false;
    return true;
}

export default Network;