import {mdiCardAccountDetails, mdiEyeOutline, mdiEyeOffOutline} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../PropertyTable";
import {Tab, Tabs} from "react-bootstrap";
import {useState} from "react";
import {secretString} from "./util";

const AccountDetail = (props) => {
    const [showToken, setShowToken] = useState(false);

    const customProperties = {
        token: row => {
            if(showToken) {
                return <span>{row.value} <Icon path={mdiEyeOffOutline} size={0.7} onClick={() => setShowToken(false)} /></span>
            } else {
                return <span>{secretString(row.value)} <Icon path={mdiEyeOutline} size={0.7} onClick={() => setShowToken(true)} /></span>
            }
        }
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