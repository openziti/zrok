import {mdiCardAccountDetails} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";

const AccountDetail = (props) => {
    return (
        <div>
            <h2><Icon path={mdiCardAccountDetails} size={2} />{" "}{props.user.email}</h2>
            <Tabs defaultActiveKey={"detail"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <PropertyTable object={props.user} />
                </Tab>
            </Tabs>
        </div>
    );
}

export default AccountDetail;