import Modal from "react-bootstrap/Modal";
import {mdiContentCopy} from "@mdi/js";
import Icon from "@mdi/react";

const Enable = (props) => {
    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);
            props.onHide();

        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    return (
        <Modal show={props.show} onHide={props.onHide} centered>
            <Modal.Header closeButton>Enable Your Environment</Modal.Header>
            <Modal.Body>
                <p>To enable your shell for zrok, use this command:</p>
                <pre>
                    $ <span id={"zrok-enable-command"}>zrok enable {props.token}</span>{' '}
                    <Icon path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                </pre>
            </Modal.Body>
        </Modal>
    );
}

export default Enable;