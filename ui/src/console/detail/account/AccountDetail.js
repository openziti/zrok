import {mdiCardAccountDetails, mdiEyeOutline, mdiEyeOffOutline} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import SecretToggle from "../../SecretToggle";

const AccountDetail = (props) => {
    const customProperties = {
        token: row => <SecretToggle secret={row.value} />
    }

    return (
        <div>
            <h2><Icon path={mdiCardAccountDetails} size={2} />{" "}{props.user.email}</h2>
            <Tabs defaultActiveKey={"detail"}>
                <Tab eventKey={"detail"} title={"Detail"}>
                    <PropertyTable object={props.user} custom={customProperties}/>
                </Tab>
            </Tabs>
        </div>
    );
}

export default AccountDetail;