import React, {useEffect, useState} from "react";
import {MetadataApi} from "../../../api/src";
import {Area, AreaChart, ResponsiveContainer} from "recharts";
import DataTable from "react-data-table-component";

const EnvironmentsTab = (props) => {
    const [detail, setDetail] = useState([]);
    const metadata = new MetadataApi()

    useEffect(() => {
        metadata.getAccountDetail()
            .then(resp => {
                setDetail(resp);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getAccountDetail()
                .then(resp => {
                    if(mounted) {
                        setDetail(resp);
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
            name: "Description",
            selector: row => row.description,
            sortable: true
        },
        {
            name: "Address",
            grow: 0.5,
            selector: row => row.address,
            sortable: true
        },
        {
            name: "Activity",
            grow: 0.5,
            cell: row => {
                return <ResponsiveContainer width={"100%"} height={"100%"}>
                    <AreaChart data={row.activity}>
                        <Area type={"basis"} dataKey={(v) => v.tx ? v.tx : 0} stroke={"#231069"} fill={"#04adef"} isAnimationActive={false} dot={false} />
                        <Area type={"basis"} dataKey={(v) => v.rx ? v.rx * -1 : 0} stroke={"#231069"} fill={"#9BF316"} isAnimationActive={false} dot={false} />
                    </AreaChart>
                </ResponsiveContainer>
            }
        }
    ];

    return (
        <div className={"zrok-datatable"}>
            <DataTable
                className={"zrok-datatable"}
                data={detail}
                columns={columns}
                defaultSortField={1}
                noDataComponent={<p>No environments in account</p>}
            />
        </div>
    );
}

export default EnvironmentsTab;