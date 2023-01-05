import PropertyTable from "../../PropertyTable";
import Icon from "@mdi/react";
import {mdiEyeOffOutline, mdiEyeOutline} from "@mdi/js";
import {secretString} from "../util";
import {useState} from "react";

const Detail = (props) => {
    const [showZId, setShowZId] = useState(false);

    const customProperties = {
        zId: row => {
            if(showZId) {
                return <span>{row.value} <Icon path={mdiEyeOffOutline} size={0.7} onClick={() => setShowZId(false)} /></span>
            } else {
                return <span>{secretString(row.value)} <Icon path={mdiEyeOutline} size={0.7} onClick={() => setShowZId(true)} /></span>
            }
        }
    }

    return (
        <PropertyTable object={props.environment} custom={customProperties} />
    );
};

export default Detail;