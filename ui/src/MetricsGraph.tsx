import {Grid2, Typography} from "@mui/material";
import {LineChart} from "@mui/x-charts";
import {useEffect, useState} from "react";
import {bytesToSize} from "./model/util.ts";
import {format} from "date-fns";

const MetricsGraph = ({ title, showTime, data }) => {
    const [rxData, setRxData] = useState([]);
    const [txData, setTxData] = useState([]);
    const [timestamps, setTimestamps] = useState([]);
    const dateFormat = showTime ? "dd-MMM H:mm" : "dd-MMM"

    useEffect(() => {
        if(data) {
            setRxData(data.map(v => v.rx ? v.rx : 0));
            setTxData(data.map(v => v.tx ? v.tx : 0 ));
            setTimestamps(data.map(v => v.timestamp));
        }
    }, [data]);

    return (
        <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
            <Typography variant="body1"><strong>{title}</strong></Typography>
            <LineChart width={560} height={200} slotProps={{ legend: {hidden: true}}} grid={{ vertical: true, horizontal: true }} series={[
                    { data: rxData, label: "rx", color: "#04adef", area: true, stack: "total", showMark: false, valueFormatter: (v) => { return bytesToSize(v as number) } },
                    { data: txData, label: "tx", color: "#9bf316", area: true, stack: "total", showMark: false, valueFormatter: (v) => { return bytesToSize(v as number) } }
                ]}
               xAxis={[{ scaleType: "time", data: timestamps, valueFormatter: (v) => { return format(new Date(v), dateFormat) } }]}
               yAxis={[{ valueFormatter: (v) => { return bytesToSize(v) } }]}
            />
        </Grid2>
    );
}

export default MetricsGraph;