import * as metadata from "../../../api/metadata";
import React, {useEffect, useState} from "react";
import {mdiShareVariant} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import ActionsTab from "./ActionsTab";
import SecretToggle from "../../SecretToggle";
import {Area, AreaChart, Line, LineChart, ResponsiveContainer, XAxis} from "recharts";

const ShareDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadata.getShareDetail(props.selection.id)
            .then(resp => {
                let detail = resp.data;
                detail.envZId = props.selection.envZId;
                setDetail(detail);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getShareDetail(props.selection.id)
                .then(resp => {
                    if(mounted) {
                        let detail = resp.data;
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
        metrics: row => (
            <ResponsiveContainer width={"100%"} height={"100%"}>
                <AreaChart data={row.value}>
                    <Area type="linearClosed" dataKey={(v) => v} stroke={"#231069"} fillOpacity={1} fill={"#04adef"} isAnimationActive={false} dot={false} />
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