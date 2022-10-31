import Icon from "@mdi/react";
import {mdiKey, mdiContentCopy} from "@mdi/js";
import Popover from "@mui/material/Popover";
import {useState} from "react";

const Token = (props) => {
    const [anchorEl, setAnchorEl] = useState(null);

    const handlePopoverClick = (event) => {
        setAnchorEl(event.currentTarget);
    };
    const handlePopoverClose = () => {
        setAnchorEl(null);
    }
    const popoverOpen = Boolean(anchorEl);
    const popoverId = popoverOpen ? 'token-popover' : undefined;

    const text = "zrok enable " + props.user.token
    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);
            console.log("copied enable command");
        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    return (
        <div>
            <button aria-describedby={popoverId} onClick={handlePopoverClick}><Icon path={mdiKey} size={0.7}/></button>
            <Popover
                id={popoverId}
                open={popoverOpen}
                anchorEl={anchorEl}
                onClose={handlePopoverClose}
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
            >
                <div className={"popover"}>
                    <h3>Enable zrok access in your shell:</h3>
                    <pre>
                        $ <span id={"zrok-enable-command"}>{text}</span> <Icon path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                    </pre>
                </div>
            </Popover>
        </div>
    );
}

export default Token;