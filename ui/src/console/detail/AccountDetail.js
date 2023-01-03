import {mdiCardAccountDetails} from "@mdi/js";
import Icon from "@mdi/react";

const AccountDetail = (props) => {
    return (
        <div>
            <h2><Icon path={mdiCardAccountDetails} size={2} />{" "}{props.user.email}</h2>
        </div>
    );
}

export default AccountDetail;