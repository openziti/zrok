import React, {useEffect, useState} from "react";
import * as account from "../../../../api/account";
import {Button, Container, Form, Row} from "react-bootstrap";
import Modal from "react-bootstrap/Modal";
import * as metadata from "../../../../api/metadata";

const validatePassword = (password, l, rc, rn, rs, spc, cb) => {
    if(password.length < l) {
        cb(false, "Entered password is too short! (" + l + " characters minimum)!");
        return;
    }
    if(rc && !/[A-Z]/.test(password)) {
        cb(false, "Entered password requires a capital letter!");
        return;
    }
    if(rn && !/\d/.test(password)) {
        cb(false, "Entered password requires a digit!");
        return;
    }
    if(rs) {
        if(!spc.split("").some(v => password.includes(v))) {
            cb(false, "Entered password requires a special character!");
            return;
        }
    }
    return cb(true, "");
}

const ChangePassword = (props) => {
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [message, setMessage] = useState('');

    const [passwordLength, setPasswordLength] = useState(8);
    const [passwordRequireCapital, setPasswordRequireCapital] = useState(true);
    const [passwordRequireNumeric, setPasswordRequireNumeric] = useState(true);
    const [passwordRequireSpecial, setPasswordRequireSpecial] = useState(true);
    const [passwordValidSpecialCharacters, setPasswordValidSpecialCharacters] = useState("");

    useEffect(() => {
        metadata.configuration().then(resp => {
            if (!resp.error) {
                setPasswordLength(resp.data.passwordRequirements.length)
                setPasswordRequireCapital(resp.data.passwordRequirements.requireCapital)
                setPasswordRequireNumeric(resp.data.passwordRequirements.requireNumeric)
                setPasswordRequireSpecial(resp.data.passwordRequirements.requireSpecial)
                setPasswordValidSpecialCharacters(resp.data.passwordRequirements.validSpecialCharacters)
            }
        }).catch(err => {
            console.log("error getting configuration", err);
        });
    }, [])

    const handleSubmit = async e => {
        e.preventDefault();

        let ok = false;
        validatePassword(newPassword,
            passwordLength,
            passwordRequireCapital,
            passwordRequireNumeric,
            passwordRequireSpecial,
            passwordValidSpecialCharacters, (isOk, msg) => { ok = isOk; setMessage(msg); })
        if(!ok) {
            return;
        }

        if(confirmPassword !== newPassword) {
            setMessage("New password and confirmation do not match!");
            return;
        }

        account.changePassword({ body: { oldPassword: oldPassword, newPassword: newPassword, email: props.user.email } })
            .then(resp => {
                if (!resp.error) {
                    console.log("resp", resp)
                    setMessage("Password successfully changed!");
                } else {
                    setMessage("Failure changing password! Is old password correct?");
                }
            }).catch(resp => {
                console.log("resp", resp)
                setMessage("Failure changing password! Is old password correct?")
            })

    }

    let hide = () => {
        props.onHide();
        setMessage("");
        setOldPassword("");
        setNewPassword("");
        setConfirmPassword("");
    }

    return (
        <Modal show={props.show} onHide={hide} size={"md"} centered>
            <Modal.Header closeButton>Change Password</Modal.Header>
            <Modal.Body>
                <Form onSubmit={handleSubmit}>
                    <Container>
                        <Form.Group controlId={"oldPassword"}>
                            <Form.Control
                                type={"password"}
                                placeholder={"Old Password"}
                                onChange={t => { setMessage(''); setOldPassword(t.target.value); }}
                                value={oldPassword}
                                style={{ marginBottom: "1em" }}
                            />
                        </Form.Group>
                        <Form.Group controlId={"newPassword"}>
                            <Form.Control
                                type={"password"}
                                placeholder={"New Password"}
                                onChange={t => { setMessage(''); setNewPassword(t.target.value); }}
                                value={newPassword}
                                style={{ marginBottom: "1em" }}
                            />
                        </Form.Group>
                        <Form.Group controlId={"confirmPassword"}>
                            <Form.Control
                                type={"password"}
                                placeholder={"Confirm Password"}
                                onChange={t => { setMessage(''); setConfirmPassword(t.target.value); }}
                                value={confirmPassword}
                            />
                        </Form.Group>
                        <Row style={{ justifyContent: "center", marginTop: "1em" }}>
                            <p style={{ color: "red" }}>{message}</p>
                        </Row>
                        <Row style={{ justifyContent: "right", marginTop: "1em" }}>
                            <Button variant={"danger"} type={"submit"}>Change Password</Button>
                        </Row>
                    </Container>
                </Form>
        </Modal.Body>
        </Modal>
    );
}

export default ChangePassword;