import {useState} from "react";
import Icon from "@mdi/react";
import {mdiEyeOffOutline, mdiEyeOutline} from "@mdi/js";

const SecretToggle = (props) => {
    const [showSecret, setShowSecret] = useState(false);

    const secretString = (s) => {
        let out = "";
        for(let i = 0; i < s.length; i++) {
            out += "*";
        }
        return out;
    }

    const toggleShow = () => setShowSecret(!showSecret);

    if(showSecret) {
        return (
            <span>{props.secret} <Icon path={mdiEyeOffOutline} size={0.7} onClick={toggleShow} /></span>
        );
    } else {
        return (
            <span>{secretString(props.secret)} <Icon path={mdiEyeOutline} size={0.7} onClick={toggleShow} /></span>
        )
    }
};

export default SecretToggle;