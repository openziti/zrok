import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";

const ShareNode = ({ data }) => {
    let shareHandle = <></>;
    if(data.accessed) {
        shareHandle = <Handle type="target" position={Position.Bottom} id="access"/>;
    }
    return (
        <>
            <Handle type="target" position={Position.Top} />
            {shareHandle}
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><ShareIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
        </>
    );
}

export default ShareNode;