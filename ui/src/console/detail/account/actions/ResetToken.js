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
        console.log("I should reset my token")
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
                    <p>You will need to update your environment file ($HOME/.zrok/environmetn.json)</p>
                    Token: <span id={"zrok-token"}>{resp.data.token}</span>{' '}
                        <Icon ref={target} path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                </div>
            ));
            setModalHeader((
                <span>Token Reset Successful</span>
            ))
        }).catch(err => {
            console.log("err", err);
        });
    }

    let hide = () => {
        setModalBody(defaultModal)
        props.onHide()
    }

    let defaultHeader = (<span>WARNING - Are you Sure?</span>)
    let defaultModal = (
        <div>
            <div>
                <div>Reseting your token will revoke access from any CLI environments.</div>
                <div>You will need to update $HOME/.zrok/environments.yml with your new token.</div>
            </div>
            <div style={{display: 'flex', alignItems:'center', justifyContent: 'center'}}>
                <Button variant={"light"} onClick={resetToken}>Reset Token</Button>
                <Button variant={"dark"} onClick={props.onHide}>Cancel</Button>
            </div>
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