import {mdiCardAccountDetails} from "@mdi/js";
import Icon from "@mdi/react";
import PropertyTable from "../PropertyTable";

const AccountDetail = (props) => {
    return (
        <div>
            <h2><Icon path={mdiCardAccountDetails} size={2} />{" "}{props.user.email}</h2>
            <PropertyTable object={props.user} />
        </div>
    );
}

export default AccountDetail;