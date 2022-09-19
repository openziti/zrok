import Icon from "@mdi/react";
import {mdiKey} from "@mdi/js";
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
                        $ zrok enable {props.user.token}
                    </pre>
                </div>
            </Popover>
        </div>
    );
}

export default Token;