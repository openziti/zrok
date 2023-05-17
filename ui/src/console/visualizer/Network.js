import {withSize} from "react-sizeme";
import {useEffect, useRef} from "react";
import {ForceGraph2D} from "react-force-graph";
import * as d3 from "d3-force-3d";
import {roundRect} from "./draw";
import {mdiShareVariant, mdiConsole, mdiAccount, mdiAlertOctagram, mdiAccessPointNetwork} from "@mdi/js";

const accountIcon = new Path2D(mdiAccount);
const environmentIcon = new Path2D(mdiConsole);
const frontendIcon = new Path2D(mdiAccessPointNetwork);
const limitIcon = new Path2D(mdiAlertOctagram);
const shareIcon = new Path2D(mdiShareVariant);

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
        let nodeColor = node.selected ? "#9BF316" : "#04adef";
        let textColor = "black";

        ctx.textBaseline = "middle";
        ctx.textAlign = "center";
        ctx.font = "6px 'Russo One'";
        let extents = ctx.measureText(node.label);
        let nodeWidth = extents.width + 10;

        ctx.fillStyle = nodeColor;
        roundRect(ctx, node.x - (nodeWidth / 2), node.y - 7, nodeWidth, 14, 1.25);
        ctx.fill();

        const nodeIcon = new Path2D();
        let xform = new DOMMatrix();
        xform.translateSelf(node.x - (nodeWidth / 2) - 6, node.y - 13);
        xform.scaleSelf(0.5, 0.5);
        switch(node.type) {
            case "share":
                nodeIcon.addPath(shareIcon, xform);
                break;
            case "environment":
                nodeIcon.addPath(environmentIcon, xform);
                break;
            case "frontend":
                nodeIcon.addPath(frontendIcon, xform);
                break;
            case "account":
                nodeIcon.addPath(accountIcon, xform);
                break;
            default:
                break;
        }
        ctx.fill(nodeIcon);
        ctx.strokeStyle = "black";
        ctx.lineWidth = 0.5;
        ctx.stroke(nodeIcon);

        if(node.limited) {
            const nodeLimitIcon = new Path2D();
            let limitXform = new DOMMatrix();
            limitXform.translateSelf(node.x + (nodeWidth / 2) - 6, node.y - 13);
            limitXform.scaleSelf(0.5, 0.5);
            nodeLimitIcon.addPath(limitIcon, limitXform);
            ctx.fillStyle = "red";
            ctx.fill(nodeLimitIcon);
            ctx.stroke(nodeLimitIcon);
        }

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
            linkWidth={(l) => l.type === "data" ? 3.0 : 1.5 }
            linkLineDash={(l) => l.type === "data" ? [3, 3] : [] }
            nodeCanvasObject={paintNode}
            backgroundColor={"linear-gradient(180deg, #0E0238 0%, #231069 100%);"}
            cooldownTicks={300}
        />
    )
}

export default withSize()(Network);