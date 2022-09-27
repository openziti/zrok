import Icon from "@mdi/react";
import {mdiContentCopy} from "@mdi/js";

const Copy = (props) => {
    function handleClick(event) {
        navigator.clipboard.writeText(props.text);
        console.log("copied", props.text);
    }

    return (
        <button onClick={handleClick}><Icon path={mdiContentCopy} size={0.5}/></button>
    );
}

export default Copy;