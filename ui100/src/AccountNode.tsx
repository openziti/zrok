import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import useMetricsStore from "./model/store.ts";
import {SparkLineChart} from "@mui/x-charts";
import {useEffect, useState} from "react";


const AccountNode = ({ data }) => {
    const environmentMetrics = useMetricsStore((state) => state.environments);
    const [sparkData, setSparkData] = useState<number[]>(Array<number>(31).fill(0));

    useEffect(() => {
        let s = new Array<number>(31);
        if(environmentMetrics) {
            environmentMetrics.forEach(env => {
                if(env.activity) {
                    env.activity.forEach((sample, i) => {
                        s[i] = s[i] + sample.rx ? sample.rx : 0;
                        s[i] = s[i] + sample.tx ? sample.tx : 0;
                    });
                }
            });
        }
        setSparkData(s);
    }, [environmentMetrics]);

    return (
        <>
            <Handle type="source" position={Position.Bottom} />
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccountIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
            <Grid2 container sx={{ flexGrow: 1, p: 0.5 }}>
                <SparkLineChart data={sparkData} height={30} width={100} colors={['#04adef']}  />
            </Grid2>
        </>
    );
}

export default AccountNode;