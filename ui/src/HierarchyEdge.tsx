import {BaseEdge, EdgeProps, getBezierPath, Position} from "@xyflow/react";

const HierarchyEdge = (props: EdgeProps) => {
    const { sourceX, sourceY, targetX, targetY, id, markerEnd } = props;

    const [edgePath] = getBezierPath({
        sourceX,
        sourceY,
        sourcePosition: Position.Bottom,
        targetX,
        targetY,
        targetPosition: Position.Top,
    });

    return <BaseEdge id={id} path={edgePath} markerEnd={markerEnd} />;
}

export default HierarchyEdge;
