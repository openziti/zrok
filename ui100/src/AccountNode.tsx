import {Handle, Position} from "@xyflow/react";
import {Grid2} from "@mui/material";
import AccountIcon from "@mui/icons-material/Person4";
import useApiConsoleStore from "./model/store.ts";
import {SparkLineChart} from "@mui/x-charts";
import {useEffect, useState} from "react";


const AccountNode = ({ data }) => {
    const environments = useApiConsoleStore((state) => state.environments);
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
            environments.forEach(env => {
                if(env.activity) {
                    env.activity.forEach((sample, i) => {
                        let v = s[i] ? s[i] : 0;
                        v += sample.rx! ? sample.rx! : 0;
                        v += sample.tx! ? sample.tx! : 0;
                        s[i] = v;
                    });
                }
            });
        }
        setSparkData(s);
    }, [environments]);

    return (
        <>
            <Handle type="source" position={Position.Bottom} />
            <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Grid2 display="flex"><AccountIcon sx={{ fontSize: 15, mr: 0.5 }}/></Grid2>
                <Grid2 display="flex">{data.label}</Grid2>
            </Grid2>
            {sparkData.find(x => x > 0) ? visibleSparkline : hiddenSparkline}
        </>
    );
}

export default AccountNode;