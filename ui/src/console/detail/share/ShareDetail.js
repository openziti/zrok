import React, {useEffect, useState} from "react";
import {mdiShareVariant} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import ActionsTab from "./ActionsTab";
import SecretToggle from "../../SecretToggle";
import {Area, AreaChart, ResponsiveContainer} from "recharts";
import MetricsTab from "./MetricsTab";
import { metadataApi } from "../../..";

const ShareDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadataApi.getShareDetail(props.selection.shrToken)
            .then(resp => {
                let detail = resp;
                detail.envZId = props.selection.envZId;
                setDetail(detail);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadataApi.getShareDetail(props.selection.shrToken)
                .then(resp => {
                    if(mounted) {
                        let detail = resp;
                        detail.envZId = props.selection.envZId;
                        setDetail(detail);
                    }
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.selection]);

    const customProperties = {
        activity: row => (
            <ResponsiveContainer width={"100%"} height={"100%"}>
                <AreaChart data={row.value}>
                    <Area type={"basis"} dataKey={(v) => v.tx ? v.tx : 0} stroke={"#231069"} fill={"#04adef"} isAnimationActive={false} dot={false} />
                    <Area type={"basis"} dataKey={(v) => v.rx ? v.rx * -1 : 0} stroke={"#231069"} fill={"#9BF316"} isAnimationActive={false} dot={false} />
                </AreaChart>
            </ResponsiveContainer>
        ),
        frontendEndpoint: row => (
            <a href={row.value} target="_">{row.value}</a>
        ),
        backendProxyEndpoint: row => {
            if(row.value.startsWith("http://") || row.value.startsWith("https://")) {
                return <a href={row.value} target="_">{row.value}</a>;
            }
            return row.value;
        },
        zId: row => <SecretToggle secret={row.value} />
    }

    if(detail) {
        return (
            <div>
                <h2><Icon path={mdiShareVariant} size={2} />{" "}{detail.backendProxyEndpoint}</h2>
                <Tabs defaultActiveKey={"detail"}>
                    <Tab eventKey={"detail"} title={"Detail"}>
                        <PropertyTable object={detail} custom={customProperties} />
                    </Tab>
                    <Tab eventKey={"metrics"} title={"Metrics"}>
                        <MetricsTab share={detail} />
                    </Tab>
                    <Tab eventKey={"actions"} title={"Actions"}>
                        <ActionsTab share={detail} />
                    </Tab>
                </Tabs>
            </div>
        );
    }
    return <></>;
}

export default ShareDetail;