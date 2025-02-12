import {Handle, Position, useUpdateNodeInternals} from "@xyflow/react";
import {Grid2} from "@mui/material";
import ShareIcon from "@mui/icons-material/Share";
import useApiConsoleStore from "./model/store.ts";
import {SparkLineChart} from "@mui/x-charts";

const ShareNode = ({ data }) => {
    const sparkdata = useApiConsoleStore((state) => state.sparkdata);
    const updateNodeInternals = useUpdateNodeInternals();

    let shareHandle = <></>;
    if(data.accessed) {
        shareHandle = <Handle type="target" position={Position.Bottom} id="access" />;
        updateNodeInternals(data.id);
    }

    const hiddenSparkline = <></>;
    const visibleSparkline = (
        <Grid2 container sx={{ flexGrow: 1, p: 0.5 }}>
            <SparkLineChart data={sparkdata.get(data.shareToken) ? sparkdata.get(data.shareToken)! : []} height={30} width={100} colors={['#04adef']}  />
        </Grid2>
    );

    return (
        <>
            <Handle type="target" position={Position.Top} />
            {shareHandle}
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><ShareIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
            {sparkdata.get(data.shareToken)?.find(x => x > 0) ? visibleSparkline : hiddenSparkline}
        </>
    );
}

export default ShareNode;