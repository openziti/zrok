import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import EnvironmentIcon from "@mui/icons-material/Computer";
import useStore from "./model/store.ts";
import {useEffect, useState} from "react";
import {SparkLineChart} from "@mui/x-charts";

const EnvironmentNode = ({ data }) => {
    const environments = useStore((state) => state.environments);
    const [sparkData, setSparkData] = useState<number[]>(Array<number>(31).fill(0));
    const hiddenSparkline = <></>;
    const visibleSparkline = (
        <Grid2 container sx={{ flexGrow: 1, p: 0.5 }}>
            <SparkLineChart data={sparkData} height={30} width={100} colors={['#04adef']}  />
        </Grid2>
    );

    useEffect(() => {
        let s = new Array<number>(31);
        if(environments) {
            let env = environments.find(env => data.envZId === env.zId);
            if(env) {
                env.activity?.forEach((sample, i) => {
                    let v = s[i] ? s[i] : 0;
                    v += sample.rx! ? sample.rx! : 0;
                    v += sample.tx! ? sample.tx! : 0;
                    s[i] = v;
                });
                setSparkData(s);
            } else {
                console.log("not found", data, environments);
            }
        }
    }, [environments]);

    return (
        <>
            <Handle type="target" position={Position.Top} />
            <Handle type="source" position={Position.Bottom} />
            <Grid2 container sx={{ flexGrow: 1, p: 1 }}>
                <Grid2 display="flex"><EnvironmentIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
            {sparkData.find(x => x > 0) ? visibleSparkline : hiddenSparkline}
        </>
    );
}

export default EnvironmentNode;