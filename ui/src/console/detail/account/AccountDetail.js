import {mdiAccountBox} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs, Tooltip} from "react-bootstrap";
import SecretToggle from "../../SecretToggle";
import React, {useEffect, useState} from "react";
import * as metadata from "../../../api/metadata";
import {Area, AreaChart, CartesianGrid, Line, LineChart, ResponsiveContainer, XAxis, YAxis} from "recharts";

const AccountDetail = (props) => {
    const customProperties = {
        token: row => <SecretToggle secret={row.value} />
    }

    return (
        <div>
            <h2><Icon path={mdiAccountBox} size={2} />{" "}{props.user.email}</h2>
            <Tabs defaultActiveKey={"detail"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <PropertyTable object={props.user} custom={customProperties}/>
                </Tab>
                <Tab eventKey={"metrics"} title={"Metrics"}>
                    <MetricsTab />
                </Tab>
            </Tabs>
        </div>
    );
}

const MetricsTab = (props) => {
    const [metrics, setMetrics] = useState({});
    const [tx, setTx] = useState(0);
    const [rx, setRx] = useState(0)

    useEffect(() => {
        metadata.getAccountMetrics()
            .then(resp => {
                setMetrics(resp.data);
            });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getAccountMetrics()
                .then(resp => {
                    if(mounted) {
                        setMetrics(resp.data);
                    }
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    useEffect(() => {
        let txAccum = 0
        let rxAccum = 0
        if(metrics.samples) {
            metrics.samples.forEach(sample => {
                txAccum += sample.tx
                rxAccum += sample.rx
            })
        }
        setTx(txAccum);
        setRx(rxAccum);
    }, [metrics])

    console.log(metrics);

    return (
        <div>
            <div>
                <h1>RX: {bytesToSize(rx)}, TX: {bytesToSize(tx)}</h1>
            </div>
            <ResponsiveContainer width={"100%"} height={300}>
                <LineChart data={metrics.samples}>
                    <CartesianGrid strokeDasharay={"3 3"} />
                    <XAxis dataKey={(v) => new Date(v.timestamp)} />
                    <YAxis />
                    <Line type={"linear"} stroke={"red"} dataKey={"rx"} activeDot={{ r: 8 }}/>
                    <Line type={"linear"} stroke={"green"} dataKey={"tx"} />
                    <Tooltip />
                </LineChart>
            </ResponsiveContainer>
        </div>
    );
}

const bytesToSize = (sz) => {
    let absSz = sz;
    if(absSz < 0) {
        absSz *= -1;
    }
    const unit = 1000
    if(absSz < unit) {
        return '' + absSz + ' B';
    }
    let div = unit
    let exp = 0
    for(let n = absSz / unit; n >= unit; n /= unit) {
        div *= unit;
        exp++;
    }

    return '' + (sz / div).toFixed(2) + "kMGTPE"[exp];
}

export default AccountDetail;