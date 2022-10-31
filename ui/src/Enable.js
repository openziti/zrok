import Icon from "@mdi/react";
import {mdiContentCopy} from "@mdi/js";

const Enable = (props) => {
    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);
            console.log("copied enable command");
        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    return <>
        <div id={"zrok-enable"}>
            <h1>Enable an Environment</h1>
            <p>To enable your shell for zrok, use this command:</p>
            <pre>
                $ <span id={"zrok-enable-command"}>zrok enable {props.token}</span> <Icon path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
            </pre>
        </div>
    </>
}

export default Enable