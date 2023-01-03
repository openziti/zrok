import * as metadata from "../../api/metadata";
import {Sparklines, SparklinesLine, SparklinesSpots} from "react-sparklines";
import {useEffect, useState} from "react";
import {mdiShareVariant} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";

const ShareDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        metadata.getServiceDetail(props.selection.id)
            .then(resp => {
               setDetail(resp.data);
            });
    }, [props.selection]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.getServiceDetail(props.selection.id)
                .then(resp => {
                    setDetail(resp.data);
                });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, [props.selection]);

    const customProperties = {
        metrics: row => <Sparklines data={row.value} limit={60} height={10}>
            <SparklinesLine color={"#3b2693"}/>
            <SparklinesSpots/>
        </Sparklines>,
        frontendEndpoint: row => <a href={row.value} target="_">{row.value}</a>,
        backendProxyEndpoint: row => {
            if(row.value.startsWith("http://") || row.value.startsWith("https://")) {
                return <a href={row.value} target="_">{row.value}</a>;
            }
            return row.value;
        }
    }

    if(detail) {
        return (
            <div>
                <h2><Icon path={mdiShareVariant} size={2} />{" "}{detail.backendProxyEndpoint}</h2>
                <Tabs defaultActiveKey={"detail"}>
                    <Tab eventKey={"detail"} title={"Detail"}>
                        <PropertyTable object={detail} custom={customProperties} />
                    </Tab>
                </Tabs>
            </div>
        );
    }
    return <></>;
}

export default ShareDetail;