import {Handle, Position, useUpdateNodeInternals} from "@xyflow/react";
import {Grid2} from "@mui/material";
import AccessIcon from "@mui/icons-material/Lan";

const AccessNode = ({ data }) => {
    const updateNodeInternals = useUpdateNodeInternals();

    let shareHandle = <></>;
    if(data.ownedShare) {
        shareHandle = <Handle type="source" position={Position.Bottom} id="share" />;
        updateNodeInternals(data.id);
    }

    return (
        <>
            <Handle type="target" position={Position.Top} />
            {shareHandle}
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccessIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.bindAddress ? data.bindAddress : data.label}</Grid2>
            </Grid2>
        </>
    );
}

export default AccessNode;