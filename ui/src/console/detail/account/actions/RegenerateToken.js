import Modal from "react-bootstrap/Modal";
import {Button, Container, Form, Row} from "react-bootstrap";
import React, {useState} from "react";
import {accountApi} from "../../../..";

const RegenerateToken = (props) => {
    const [confirmEmail, setConfirmEmail] = useState('');
    const [message, setMessage] = useState('');


    const hide = () => {
        props.onHide();
        setConfirmEmail('');
        setMessage('');
    };

    const handleSubmit = (event) => {
        event.preventDefault();

        if(confirmEmail !== props.user.email) {
            setMessage("Email address confirmation does not match!");
            return;
        }

        accountApi.resetToken({body: {emailAddress: props.user.email}})
            .then(resp => {
                console.log(resp);
                let user = JSON.parse(localStorage.getItem('user'));
                localStorage.setItem('user', JSON.stringify({
                    email: user.email,
                    token: resp.token
                }));
                document.dispatchEvent(new Event('storage'));
                setMessage("Your new account token is: " + resp.token);
            }).catch(err => {
                setMessage("Account token regeneration failed!");
                console.log("account token regeneration failed", err);
            });
    };

    return (
        <Modal show={props.show} onHide={hide} size={"md"} centered>
            <Modal.Header closeButton>Are you very sure?</Modal.Header>
            <Modal.Body>
                <Form onSubmit={handleSubmit}>
                    <Container>
                        <p>
                            Did you read the warning on the previous screen? This action will reset all of your active
                            environments and shares!
                        </p>
                        <p>
                            You will need to update each of
                            your <code> &#36;&#123;HOME&#125;/.zrok/environments.yml</code> files
                            with your new token to allow them to continue working!
                        </p>
                        <p>
                            Hit <code> Escape </code> or click the 'X' to abort!
                        </p>
                        <Form.Group controlId={"confirmEmail"}>
                            <Form.Control
                                placeholder={"Confirm Your Email Address"}
                                onChange={t => {
                                    setMessage('');
                                    setConfirmEmail(t.target.value);
                                }}
                                value={confirmEmail}
                                style={{marginBottom: "1em"}}
                            />
                        </Form.Group>
                        <Row style={{ justifyContent: "center", marginTop: "1em" }}>
                            <p style={{ color: "red" }}>{message}</p>
                        </Row>
                        <Row style={{ justifyContent: "right", marginTop: "1em" }}>
                            <Button variant={"danger"} type={"submit"}>Regenerate Account Token</Button>
                        </Row>
                    </Container>
                </Form>
            </Modal.Body>
        </Modal>
    );
};

export default RegenerateToken;