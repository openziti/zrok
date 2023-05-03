import {withSize} from "react-sizeme";
import {useEffect, useRef} from "react";
import {ForceGraph2D} from "react-force-graph";
import * as d3 from "d3-force-3d";
import {roundRect} from "./draw";

const Network = (props) => {
    const targetRef = useRef();
    if(props.setRef != null) {
        props.setRef(targetRef);
    }

    useEffect(() => {
        const fg = targetRef.current;
        fg.d3Force('collide', d3.forceCollide().radius(35));
    }, []);

    const paintNode = (node, ctx) => {
        let nodeColor = node.selected ? "#04adef" : "#9BF316";
        let textColor = node.selected ? "white" : "black";

        ctx.textBaseline = "middle";
        ctx.textAlign = "center";
        ctx.font = "6px 'Russo One'";
        let extents = ctx.measureText(node.label);
        let nodeWidth = extents.width + 10;

        ctx.fillStyle = nodeColor;
        roundRect(ctx, node.x - (nodeWidth / 2), node.y - 7, nodeWidth, 14, 1.25);
        ctx.fill();

        ctx.fillStyle = textColor;
        ctx.fillText(node.label, node.x, node.y);
    }

    const nodeClicked = (node) => {
        props.setSelection(node);
    }

    return (
        <ForceGraph2D
            ref={targetRef}
            graphData={props.networkGraph}
            width={props.size.width}
            height={800}
            onNodeClick={nodeClicked}
            linkOpacity={.75}
            linkWidth={1.5}
            nodeCanvasObject={paintNode}
            backgroundColor={"linear-gradient(180deg, #0E0238 0%, #231069 100%);"}
            cooldownTicks={300}
        />
    )
}

export default withSize()(Network);