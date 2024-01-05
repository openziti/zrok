import {mdiAccessPointNetwork} from "@mdi/js";
import Icon from "@mdi/react";
import {useEffect, useState} from "react";
import {getFrontendDetail} from "../../../api/metadata";
import {Tab, Tabs} from "react-bootstrap";
import DetailTab from "./DetailTab";
import ActionsTab from "./ActionsTab";

const AccessDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        getFrontendDetail(props.selection.feId)
            .then(resp => {
                setDetail(resp.data);
            });
    }, [props.selection]);

    return (
        <div>
            <h2><Icon path={mdiAccessPointNetwork} size={2} />{" "}{detail.token}</h2>
            <Tabs defaultActiveKey={"detail"} className={"mb-3"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <DetailTab frontend={detail} />
                </Tab>
                <Tab eventKey={"actions"} title={"Actions"}>
                    <ActionsTab frontend={detail} />
                </Tab>
            </Tabs>
        </div>
    );
}

export default AccessDetail;