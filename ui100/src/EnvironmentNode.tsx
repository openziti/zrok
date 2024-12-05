import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";

const EnvironmentNode = ({ data }) => {
    let shareHandle = <Handle type="source" position={Position.Bottom} />;
    if(data.empty) {
        shareHandle = <></>;
    }
    return (
        <>
            <Handle type="target" position={Position.Top} />
            {shareHandle}
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
        </>
    );
}

export default EnvironmentNode;