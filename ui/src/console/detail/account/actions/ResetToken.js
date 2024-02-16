import React, {useRef, useState} from "react";
import Modal from "react-bootstrap/Modal";
import {mdiContentCopy} from "@mdi/js";
import Icon from "@mdi/react";
import { Button, Overlay, Tooltip } from "react-bootstrap";
import * as account from "../../../../api/account";

const ResetToken = (props) => {
    const target = useRef(null);
    const [showTooltip, setShowTooltip] = useState(false);

    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-token").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);

            setShowTooltip(true);
            setTimeout(() => setShowTooltip(false), 1000);

        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    let resetToken = () => {
        account.resetToken({ body: { "emailAddress": props.user.email } }).then(resp => {
            console.log(resp)
            let user = JSON.parse(localStorage.getItem('user'))
            localStorage.setItem('user', JSON.stringify({
                "email": user.email,
                "token": resp.data.token
            }));
            document.dispatchEvent(new Event('storage'))
            setModalBody((
                <div>
                    <p>
                        You will need to update your environment files <code> &#36;&#123;HOME&#125;/.zrok/environment.json </code>
                        with the new <code> zrok_token </code>.
                    </p>
                    <p>
                        Your new <code> zrok_token </code> is: <code><span id={"zrok-token"}>{resp.data.token}</span></code>{' '}
                        <Icon ref={target} path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                    </p>

                </div>
            ));
            setModalHeader((
                <span>Account Token Regenerated!</span>
            ))
        }).catch(err => {
            console.log("err", err);
        });
    }

    let hide = () => {
        setModalHeader(defaultHeader)
        setModalBody(defaultModal)
        props.onHide()
    }

    let defaultHeader = (<span>Are you sure?</span>)
    let defaultModal = (
        <div>
            <p>Did you read the warning on the previous screen? This action will reset all of your active environments and shares!</p>
            <p>You will need to update each of your <code> &#36;&#123;HOME&#125;/.zrok/environments.yml</code> files with your new token!</p>
            <p align={"right"}>
                <Button onClick={props.onHide}>Cancel</Button>
                <Button variant={"danger"} onClick={resetToken}>Regenerate Token</Button>
            </p>
        </div>
    );

    const [modalBody, setModalBody] = useState(defaultModal);
    const [modalHeader, setModalHeader] = useState(defaultHeader);

    return (
        <div>
            <Modal show={props.show} onHide={hide} centered>
                <Modal.Header closeButton>{modalHeader}</Modal.Header>
                <Modal.Body>
                    {modalBody}
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
    )
}

export default ResetToken;