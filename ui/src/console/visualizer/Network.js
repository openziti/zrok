import {withSize} from "react-sizeme";
import {useEffect, useRef} from "react";
import {ForceGraph2D} from "react-force-graph";
import * as d3 from "d3-force-3d";

const Network = (props) => {
    const targetRef = useRef();
    if(props.setRef != null) {
        props.setRef(targetRef);
    }

    useEffect(() => {
        const fg = targetRef.current;
        fg.d3Force('collide', d3.forceCollide().radius(30));
    }, []);

    const paintNode = (node, ctx) => {
        let nodeColor = "#777";
        let textColor = "white";
        switch(node.type) {
            case "service":
                nodeColor = "#7e67e2";
                break;
        }

        ctx.textBaseline = "middle";
        ctx.textAlign = "center";
        ctx.font = "6px 'JetBrains Mono'";
        let extents = ctx.measureText(node.label);
        let nodeWidth = extents.width + 5;

        ctx.fillStyle = nodeColor;
        ctx.fillRect(node.x - (nodeWidth / 2), node.y - 7, nodeWidth, 14);

        ctx.fillStyle = textColor;
        ctx.fillText(node.label, node.x, node.y);
    }

    const nodeClicked = (node) => {
        console.log("node clicked", node.label);
    }

    return (
        <ForceGraph2D
            ref={targetRef}
            graphData={props.networkGraph}
            width={props.size.width}
            height={500}
            onNodeClick={nodeClicked}
            linkOpacity={.75}
            linkWidth={1.5}
            nodeCanvasObject={paintNode}
            backgroundColor={"#3b2693"}
            cooldownTicks={100}
        />
    )
}

export default withSize()(Network);