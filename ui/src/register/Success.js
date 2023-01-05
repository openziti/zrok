import Icon from "@mdi/react";
import {mdiContentCopy} from "@mdi/js";
import {Container, Overlay, Row, Tooltip} from "react-bootstrap";
import React, {useRef, useState} from "react";

const Success = (props) => {
    const [showTooltip, setShowTooltip] = useState(false);
    const target = useRef(null);

    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);
            console.log("copied enable command");

            setShowTooltip(true);
            setTimeout(() => setShowTooltip(false), 1000);

        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    // clear local storage on new account registration success.
    localStorage.clear();

    return (
        <Container fluid>
            <Row>
                <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
            </Row>
            <Row>
                <h1>Welcome to zrok!</h1>
            </Row>
            <Row className={"fullscreen-body"}>
                <Container className={"fullscreen-form"}>
                    <Row>
                        <p>You can proceed to the <a href={"/"}>zrok web portal</a> and log in with your email and password.</p>
                    </Row>
                    <Row>
                        <p>To enable your shell for zrok, use this command:</p>
                    </Row>
                    <Row>
                        <pre>
                           $ <span id={"zrok-enable-command"}>zrok enable {props.token}</span> <Icon ref={target} path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
                        </pre>
                        <Overlay target={target.current} show={showTooltip} placement={"bottom"}>
                            {(props) => (
                                <Tooltip id={"copy-tooltip"} {...props}>
                                    Copied!
                                </Tooltip>
                            )}
                        </Overlay>
                    </Row>
                </Container>
            </Row>
        </Container>
    )
};

export default Success;