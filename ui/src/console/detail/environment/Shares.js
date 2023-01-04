import * as metadata from "../../../api/metadata";
import {useEffect, useState} from "react";
import DataTable from 'react-data-table-component';
import {Sparklines, SparklinesLine, SparklinesSpots} from "react-sparklines";
import {mdiConsoleNetwork} from "@mdi/js";
import Icon from "@mdi/react";

const Shares = (props) => {
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
            hide: "md"
        },
        {
            name: "Backend",
            selector: row => row.backendProxyEndpoint,
            sortable: true,
        },
        {
            name: "Share Mode",
            selector: row => row.shareMode,
            hide: "md"
        },
        {
            name: "Token",
            selector: row => row.token,
            sortable: true,
            hide: "md"
        },
        {
            name: "Activity",
            cell: row => {
                return <Sparklines data={row.metrics} height={20} limit={60}><SparklinesLine color={"#3b2693"}/><SparklinesSpots/></Sparklines>;
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

export default Shares;