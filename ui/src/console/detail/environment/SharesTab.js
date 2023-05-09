import * as metadata from "../../../api/metadata";
import React, {useEffect, useState} from "react";
import DataTable from 'react-data-table-component';
import {Area, AreaChart, ResponsiveContainer} from "recharts";

const SharesTab = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadata.getEnvironmentDetail(props.selection.id)
            .then(resp => {
                setDetail(resp.data);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getEnvironmentDetail(props.selection.id)
                .then(resp => {
                    if(mounted) {
                        setDetail(resp.data);
                    }
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.selection]);

    const columns = [
        {
            name: "Frontend",
            selector: row => <a href={row.frontendEndpoint} target={"_"}>{row.frontendEndpoint}</a>,
            sortable: true,
        },
        {
            name: "Backend",
            selector: row => row.backendProxyEndpoint,
            sortable: true,
            hide: "md"
        },
        {
            name: "Activity",
            cell: row => {
                return <ResponsiveContainer width={"100%"} height={"100%"}>
                    <AreaChart data={row.metrics}>
                        <Area type="basis" dataKey={(v) => v} stroke={"#777"} fillOpacity={0.5} fill={"#04adef"} isAnimationActive={false} dot={false} />
                    </AreaChart>
                </ResponsiveContainer>
            }
        }
    ];

    if(detail.environment) {
        return (
            <div className={"zrok-datatable"}>
                <DataTable
                    className={"zrok-datatable"}
                    data={detail.shares}
                    columns={columns}
                    defaultSortField={1}
                    noDataComponent={<p>No shares in environment</p>}
                />
            </div>
        );
    }
    return <></>;
}

export default SharesTab;