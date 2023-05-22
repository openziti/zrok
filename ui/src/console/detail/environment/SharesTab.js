import * as metadata from "../../../api/metadata";
import React, {useEffect, useState} from "react";
import DataTable from 'react-data-table-component';
import {Area, AreaChart, ResponsiveContainer} from "recharts";

const SharesTab = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadata.getEnvironmentDetail(props.selection.envZId)
            .then(resp => {
                setDetail(resp.data);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getEnvironmentDetail(props.selection.envZId)
                .then(resp => {
                    if(mounted) {
                        setDetail(resp.data);
                    }
                });
        }, 5000);
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
            grow: 0.5,
            selector: row => row.backendProxyEndpoint,
            sortable: true,
            hide: "lg"
        },
        {
            name: "Activity",
            grow: 0.5,
            cell: row => {
                return <ResponsiveContainer width={"100%"} height={"100%"}>
                    <AreaChart data={row.activity}>
                        <Area type={"basis"} dataKey={(v) => v.rx ? v.rx : 0} stroke={"#231069"} fill={"#04adef"} isAnimationActive={false} dot={false} />
                        <Area type={"basis"} dataKey={(v) => v.tx ? v.tx * -1 : 0} stroke={"#231069"} fill={"#9BF316"} isAnimationActive={false} dot={false} />
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