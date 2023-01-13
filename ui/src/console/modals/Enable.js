import Modal from "react-bootstrap/Modal";
import {mdiContentCopy} from "@mdi/js";
import Icon from "@mdi/react";
import {useRef, useState} from "react";
import {Overlay, Tooltip} from "react-bootstrap";

const Enable = (props) => {
    const target = useRef(null);
    const [showTooltip, setShowTooltip] = useState(false);

    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);

            setShowTooltip(true);
            setTimeout(() => setShowTooltip(false), 1000);

        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    return (
        <div>
            <Modal show={props.show} onHide={props.onHide} centered>
                <Modal.Header closeButton>Enable Your Environment</Modal.Header>
                <Modal.Body>
                    <p>To enable your shell for zrok, use this command:</p>
                    <pre>
                    $ <span id={"zrok-enable-command"}>zrok enable {props.token}</span>{' '}
                        <Icon ref={target} path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                </pre>
                </Modal.Body>
            </Modal>
            <Overlay target={target.current} show={showTooltip} placement={"bottom"}>
                {(props) => (
                    <Tooltip id={"copy-tooltip"} {...props}>
                        Copied!
                    </Tooltip>
                )}
            </Overlay>
        </div>
    );
}

export default Enable;