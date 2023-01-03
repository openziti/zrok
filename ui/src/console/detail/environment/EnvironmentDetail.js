import {Tab, Tabs} from "react-bootstrap";
import Activity from "./Activity";
import {useEffect, useState} from "react";
import Icon from "@mdi/react";
import {mdiConsoleNetwork} from "@mdi/js";
import {getEnvironmentDetail} from "../../../api/metadata";
import Detail from "./Detail";

const EnvironmentDetail = (props) => {
    const [detail, setDetail] = useState({});

    useEffect(() => {
        getEnvironmentDetail(props.selection.id)
            .then(resp => {
                setDetail(resp.data);
            });
    }, [props.selection]);

    if(detail.environment) {
        return (
            <div>
                <h2><Icon path={mdiConsoleNetwork} size={2} />{" "}{detail.environment.description}</h2>
                <Tabs defaultActiveKey={"activity"} className={"mb-3"}>
                    <Tab eventKey={"activity"} title={"Activity"}>
                        <Activity selection={props.selection} />
                    </Tab>
                    <Tab eventKey={"detail"} title={"Detail"}>
                        <Detail environment={detail.environment} />
                    </Tab>
                </Tabs>
            </div>
        );
    }
    return <></>;
};

export default EnvironmentDetail;