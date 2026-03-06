import {BaseEdge, EdgeProps, getSmoothStepPath, Position} from "@xyflow/react";

const AccessEdge = (props: EdgeProps) => {
    const { sourceX, sourceY, targetX, targetY, id, markerEnd, data } = props;
    const laneIndex = (data as any)?.laneIndex ?? 0;
    const offset = 25 + laneIndex * 15;

    const [edgePath] = getSmoothStepPath({
        sourceX,
        sourceY,
        sourcePosition: Position.Bottom,
        targetX,
        targetY,
        targetPosition: Position.Bottom,
        borderRadius: 8,
        offset,
    });

    return <BaseEdge id={id} path={edgePath} markerEnd={markerEnd}
        style={{ strokeDasharray: "8 4", strokeWidth: 1.5 }} />;
}

export default AccessEdge;