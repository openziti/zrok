import {Grid2, Typography} from "@mui/material";
import {Area, AreaChart, CartesianGrid, XAxis, YAxis} from "recharts";
import moment from "moment";
import {bytesToSize} from "./model/util.ts";

const MetricsGraph = ({ title, data }) => {
    return (
        <Grid2 container sx={{ flexGrow: 1, p: 1 }} alignItems="center">
                <Typography variant="body1"><strong>{title}</strong></Typography>
                <AreaChart data={data} width={600} height={150}>
                    <CartesianGrid strokeDasharay={"3 3"} />
                    <XAxis dataKey={(v) => v.timestamp} scale={"time"} tickFormatter={(v) => moment(v).format("MMM DD") } style={{ fontSize: '75%', fontFamily: "Poppins" }}/>
                    <YAxis tickFormatter={(v) => bytesToSize(v)} style={{ fontSize: '75%', fontFamily: "Poppins" }}/>
                    <Area type={"basis"} stroke={"#231069"} fill={"#04adef"} dataKey={(v) => v.tx ? v.tx : 0} stackId={"1"} />
                    <Area type={"basis"} stroke={"#231069"} fill={"#9BF316"} dataKey={(v) => v.rx ? v.rx : 0} stackId={"1"} />
                </AreaChart>
        </Grid2>
    );
}

export default MetricsGraph;