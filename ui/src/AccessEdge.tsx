import {BaseEdge, EdgeProps} from "@xyflow/react";

const AccessEdge = (props: EdgeProps) => {
    const { sourceX, sourceY, targetX, targetY, id, markerEnd } = props;
    const edgePath = `M ${sourceX} ${sourceY} ` +
        `L ${sourceX} ${sourceY + 20} ` +
        `L ${sourceX + (targetX - sourceX) / 2} ${sourceY + 50 + ((sourceX - targetX) * .05) + (targetY - sourceY) / 2} ` +
        `L ${targetX} ${targetY + 20} ` +
        `L ${targetX} ${targetY}`;

    return <BaseEdge path={edgePath} markerEnd={markerEnd} />;
}

export default AccessEdge;