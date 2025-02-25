import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import {SparkLineChart} from "@mui/x-charts";
import useApiConsoleStore from "./model/store.ts";

const EnvironmentNode = ({ data }) => {
    const sparkdata = useApiConsoleStore((state) => state.sparkdata);

    const hiddenSparkline = <></>;
    const visibleSparkline = (
        <Grid2 container sx={{ flexGrow: 1, p: 0.5 }}>
            <SparkLineChart data={sparkdata.get(data.envZId) ? sparkdata.get(data.envZId)! : []} height={30} width={100} colors={['#04adef']}  />
        </Grid2>
    );

    return (
        <>
            <Handle type="target" position={Position.Top} />
            <Handle type="source" position={Position.Bottom} />
            <Grid2 container sx={{ flexGrow: 1, p: 1 }}>
                <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
            {sparkdata.get(data.envZId)?.find(x => x > 0) ? visibleSparkline : hiddenSparkline}
        </>
    );
}

export default EnvironmentNode;