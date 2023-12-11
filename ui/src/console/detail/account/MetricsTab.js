import React, {useEffect, useState} from "react";
import {buildMetrics} from "../../metrics/util";
import {MetadataApi} from "../../../api/src";
import MetricsView from "../../metrics/MetricsView";

const MetricsTab = () => {
    const [metrics30, setMetrics30] = useState(buildMetrics([]));
    const [metrics7, setMetrics7] = useState(buildMetrics([]));
    const [metrics1, setMetrics1] = useState(buildMetrics([]));
    const metadata = new MetadataApi()

    useEffect(() => {
        metadata.getAccountMetrics()
            .then(resp => {
                setMetrics30(buildMetrics(resp));
            });
        metadata.getAccountMetrics({duration: "168h"})
            .then(resp => {
                setMetrics7(buildMetrics(resp));
            });
        metadata.getAccountMetrics({duration: "24h"})
            .then(resp => {
                setMetrics1(buildMetrics(resp));
            });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getAccountMetrics()
                .then(resp => {
                    if(mounted) {
                        setMetrics30(buildMetrics(resp));
                    }
                });
            metadata.getAccountMetrics({duration: "168h"})
                .then(resp => {
                    setMetrics7(buildMetrics(resp));
                });
            metadata.getAccountMetrics({duration: "24h"})
                .then(resp => {
                    setMetrics1(buildMetrics(resp));
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

export default MetricsTab;