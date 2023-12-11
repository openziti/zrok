import {mdiAccessPointNetwork} from "@mdi/js";
import Icon from "@mdi/react";
import {useEffect, useState} from "react";
import {MetadataApi} from "../../../api/src";
import {Tab, Tabs} from "react-bootstrap";
import DetailTab from "./DetailTab";

const AccessDetail = (props) => {
    const [detail, setDetail] = useState({});
    const metadata = new MetadataApi()

    useEffect(() => {
        metadata.getFrontendDetail(props.selection.feId)
            .then(resp => {
                setDetail(resp);
            });
    }, [props.selection]);

    return (
        <div>
            <h2><Icon path={mdiAccessPointNetwork} size={2} />{" "}{detail.shrToken} ({detail.id})</h2>
            <Tabs defaultActiveKey={"detail"} className={"mb-3"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <DetailTab frontend={detail} />
                </Tab>
            </Tabs>
        </div>
    );
}

export default AccessDetail;