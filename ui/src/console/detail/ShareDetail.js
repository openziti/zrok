import * as metadata from "../../api/metadata";
import {Sparklines, SparklinesLine, SparklinesSpots} from "react-sparklines";
import {useEffect, useState} from "react";
import {mdiShareVariant} from "@mdi/js";
import Icon from "@mdi/react";
import {Tab, Tabs} from "react-bootstrap";
import PropertyTable from "../PropertyTable";

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

    if(detail) {
        return (
            <div>
                <h2><Icon path={mdiShareVariant} size={2} />{" "}{detail.backendProxyEndpoint}</h2>
                <Tabs defaultActiveKey={"activity"}>
                    <Tab eventKey={"details"} className={"mb-3"}>
                        <h3>Share Details</h3>
                    </Tab>
                    <Tab eventKey={"activity"}>
                        <div className={"zrok-big-sparkline"}>
                            <PropertyTable object={detail} />
                            <Sparklines data={detail.metrics} limit={60} height={20}>
                                <SparklinesLine color={"#3b2693"} />
                                <SparklinesSpots />
                            </Sparklines>
                        </div>
                    </Tab>
                </Tabs>
            </div>
        );
    }
    return <></>;
}

export default ShareDetail;