import {mdiAccountBox} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import SecretToggle from "../../SecretToggle";
import React, {useEffect, useState} from "react";
import * as metadata from "../../../api/metadata";
import {buildMetrics} from "../../metrics/util";
import MetricsView from "../../metrics/MetricsView";

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

const MetricsTab = () => {
    const [metrics30, setMetrics30] = useState(buildMetrics([]));
    const [metrics7, setMetrics7] = useState(buildMetrics([]));
    const [metrics1, setMetrics1] = useState(buildMetrics([]));

    useEffect(() => {
        metadata.getAccountMetrics()
            .then(resp => {
                setMetrics30(buildMetrics(resp.data));
            });
        metadata.getAccountMetrics({duration: "168h"})
            .then(resp => {
                setMetrics7(buildMetrics(resp.data));
            });
        metadata.getAccountMetrics({duration: "24h"})
            .then(resp => {
                setMetrics1(buildMetrics(resp.data));
            });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getAccountMetrics()
                .then(resp => {
                    if(mounted) {
                        setMetrics30(buildMetrics(resp.data));
                    }
                });
            metadata.getAccountMetrics({duration: "168h"})
                .then(resp => {
                    setMetrics7(buildMetrics(resp.data));
                });
            metadata.getAccountMetrics({duration: "24h"})
                .then(resp => {
                    setMetrics1(buildMetrics(resp.data));
                });
        }, 5000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    return (
        <MetricsView metrics30={metrics30} metrics7={metrics7} metrics1={metrics1} />
    );
}

export default AccountDetail;