import {BaseEdge, EdgeProps} from "@xyflow/react";

const AccessEdge = (props: EdgeProps) => {
    const { sourceX, sourceY, targetX, targetY, id, markerEnd } = props;
    const edgePath = `M ${sourceX} ${sourceY} L ${sourceX} ${sourceY + 20} L ${targetX} ${targetY + 20} L ${targetX} ${targetY}`;

    return <BaseEdge path={edgePath} markerEnd={markerEnd} />;
}

export default AccessEdge;