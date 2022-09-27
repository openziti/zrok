import Icon from "@mdi/react";
import {mdiKey} from "@mdi/js";
import Popover from "@mui/material/Popover";
import {useState} from "react";
import Copy from "./Copy";

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
                        {text} <Copy text={text}/>
                    </pre>
                </div>
            </Popover>
        </div>
    );
}

export default Token;