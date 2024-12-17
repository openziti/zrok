import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";

const AccountNode = ({ data }) => {
    return (
        <>
            <Handle type="source" position={Position.Bottom} />
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccountIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
        </>
    );
}

export default AccountNode;