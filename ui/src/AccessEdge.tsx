import {BaseEdge, Edge, EdgeProps} from "@xyflow/react";

interface AccessEdgeData {
    laneIndex?: number;
    laneCount?: number;
    laneBaseY?: number;
}

const AccessEdge = (props: EdgeProps<Edge<AccessEdgeData>>) => {
    const { sourceX, sourceY, targetX, targetY, id, markerEnd, data } = props;
    const laneIndex = data?.laneIndex ?? 0;
    const laneBaseY = data?.laneBaseY ?? Math.max(sourceY, targetY);

    const r = 8;
    const laneY = laneBaseY + 25 + laneIndex * 15;
    const dx = targetX !== sourceX ? Math.sign(targetX - sourceX) : 1;

    const edgePath = [
        `M ${sourceX},${sourceY}`,
        `L ${sourceX},${laneY - r}`,
        `Q ${sourceX},${laneY} ${sourceX + dx * r},${laneY}`,
        `L ${targetX - dx * r},${laneY}`,
        `Q ${targetX},${laneY} ${targetX},${laneY - r}`,
        `L ${targetX},${targetY}`,
    ].join(" ");

    return <BaseEdge id={id} path={edgePath} markerEnd={markerEnd}
        style={{ strokeDasharray: "8 4", strokeWidth: 1.5 }} />;
}

export default AccessEdge;