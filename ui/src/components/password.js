import React, {useEffect, useState, Fragment} from "react";
import {Button, Container, Form, Row} from "react-bootstrap";

const PasswordForm = (props) => {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');

    const passwordMismatchMessage = <h2 className={"errorMessage"}>Entered passwords do not match!</h2>
    const passwordTooShortMessage = <h2 className={"errorMessage"}>Entered password too short! ({props.passwordLength} characters, minimum)</h2>
    const passwordRequiresCapitalMessage = <h2 className={"errorMessage"}>Entered password requires a capital letter!</h2>
    const passwordRequiresNumericMessage = <h2 className={"errorMessage"}>Entered password requires a digit!</h2>
    const passwordRequiresSpecialMessage = <h2 className={"errorMessage"}>Entered password requires a special character! ({props.passwordValidSpecialCharacters.split("").join(" ")})</h2>

    useEffect(() => {
        if (confirm === "" && password === "") {
            return
        }
        if (confirm.length < props.passwordLength) {
            props.setMessage(passwordTooShortMessage)
            return;
        }
        if (props.passwordRequireCapital && !/[A-Z]/.test(confirm)) {
            props.setMessage(passwordRequiresCapitalMessage)
            return;
        }
        if (props.passwordRequireNumeric && !/\d/.test(confirm)) {
            props.setMessage(passwordRequiresNumericMessage)
            return;
        }
        if (props.passwordRequireSpecial) {
            if (!props.passwordValidSpecialCharacters.split("").some(v => confirm.includes(v))) {
                props.setMessage(passwordRequiresSpecialMessage)
                return;
            }
        }
        if (confirm != password) {
            props.setMessage(passwordMismatchMessage)
            return;
        }
        props.setParentPassword(password)
    }, [password, confirm])

    return (
        <Fragment>
            {
                (props.passwordLength > 0 || props.passwordRequireCapital || props.passwordRequireNumeric || props.passwordRequireSpecial) &&
                <h2>Password Requirements</h2>
            }
            {
                props.passwordLength > 0 &&
                <Row>Minimum Length of {props.passwordLength} </Row>
            }
            {
                props.passwordRequireCapital &&
                <Row>Requires at least 1 Capital Letter</Row>
            }
            {
                props.passwordRequireNumeric &&
                <Row>Requires at least 1 Digit</Row>
            }
            {
                props.passwordRequireSpecial &&
                <Fragment>
                    <Row>Requires at least 1 Special Character</Row>
                    <Row>{props.passwordValidSpecialCharacters.split("").join(" ")}</Row>
                </Fragment>
            }
            <Container className={"fullscreen-body"}>
                <Form.Group controlId={"password"}>
                    <Form.Control
                        type={"password"}
                        placeholder={"Set Password"}
                        onChange={t => { props.setMessage(null); setPassword(t.target.value);}}
                        value={password}
                    />
                </Form.Group>

                <Form.Group controlId={"confirm"}>
                    <Form.Control
                        type={"password"}
                        placeholder={"Confirm Password"}
                        onChange={t => { props.setMessage(null); setConfirm(t.target.value);}}
                        value={confirm}
                    />
                </Form.Group>
            </Container>
        </Fragment>
    )
};

PasswordForm.defaultProps = {
    passwordLength: 0,
    passwordRequireCapital: false,
    passwordRequireNumeric: false,
    passwordRequireSpecial: false,
    passwordValidSpecialCharacters: ""
}

export default PasswordForm;