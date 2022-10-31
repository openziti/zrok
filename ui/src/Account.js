import {useEffect, useState} from "react";
import {useReactFlow} from "react-flow-renderer";
import Icon from "@mdi/react";
import {mdiAccessPointNetwork, mdiDesktopClassic} from "@mdi/js";
import dagre from "dagre";
import * as metadata from "./api/metadata";
import Network from "./Network";
import Enable from "./Enable";

const Account = (props) => {
    const [mode, setMode] = useState(<></>);
    const reactFlow = useReactFlow();

    useEffect(() => {
        let mounted = true
        metadata.overview().then(resp => {
            if(mounted) {
                let overview = resp.data
                let g = buildGraph(resp.data)
                let nodes = getLayout(g)
                let edges = g.edges

                if(resp.data.length > 0) {
                    setMode(<Network
                        nodes={nodes}
                        edges={edges}
                        overview={overview}
                    />)
                    reactFlow.fitView({maxZoom: 1})
                } else {
                    setMode(<Enable token={props.user.token}/>)
                }
            }
        });
    }, [])

    useEffect(() => {
        let mounted = true
        let interval = setInterval(() => {
            metadata.overview().then(resp => {
                if(mounted) {
                    let overview = resp.data
                    let g = buildGraph(resp.data)
                    let nodes = getLayout(g)
                    let edges = g.edges

                    if(resp.data.length > 0) {
                        setMode(<Network
                            nodes={nodes}
                            edges={edges}
                            overview={overview}
                        />)
                        reactFlow.fitView({maxZoom: 1})
                    } else {
                        setMode(<Enable token={props.user.token}/>)
                    }
                }
            })
        }, 1000)
        return () => {
            mounted = false
            clearInterval(interval)
        }
    }, [])

    return <>{mode}</>
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
                out.nodes.push({
                    id: '' + id,
                    data: {label: <div><Icon path={mdiAccessPointNetwork} size={0.75} className={"flowNode"}/> { item.name }</div>},
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
            });
        }
    });
    return out
}

const nodeWidth = 100;
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

export default Account