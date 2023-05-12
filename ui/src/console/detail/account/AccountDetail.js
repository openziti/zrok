import {mdiAccountBox} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import SecretToggle from "../../SecretToggle";
import React from "react";
import MetricsTab from "./MetricsTab";
import EnvironmentsTab from "./EnvironmentsTab";

const AccountDetail = (props) => {
    const customProperties = {
        token: row => <SecretToggle secret={row.value} />
    }

    return (
        <div>
            <h2><Icon path={mdiAccountBox} size={2} />{" "}{props.user.email}</h2>
            <Tabs defaultActiveKey={"environments"}>
                <Tab eventKey={"environments"} title={"Environments"}>
                    <EnvironmentsTab />
                </Tab>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <PropertyTable object={props.user} custom={customProperties}/>
                </Tab>
                <Tab eventKey={"metrics"} title={"Metrics"}>
                    <MetricsTab />
                </Tab>
            </Tabs>
        </div>
    );
}



export default AccountDetail;