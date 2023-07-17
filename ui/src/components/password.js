import React, {useEffect, useState, Fragment} from "react";
import {Container, Form, Row} from "react-bootstrap";

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
        if (password.length < props.passwordLength) {
            props.setMessage(passwordTooShortMessage)
            return;
        }
        if (props.passwordRequireCapital && !/[A-Z]/.test(password)) {
            props.setMessage(passwordRequiresCapitalMessage)
            return;
        }
        if (props.passwordRequireNumeric && !/\d/.test(password)) {
            props.setMessage(passwordRequiresNumericMessage)
            return;
        }
        if (props.passwordRequireSpecial) {
            if (!props.passwordValidSpecialCharacters.split("").some(v => password.includes(v))) {
                props.setMessage(passwordRequiresSpecialMessage)
                return;
            }
        }
        if (confirm !== password) {
            props.setMessage(passwordMismatchMessage)
            return;
        }
        props.setParentPassword(password)
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [password, confirm])

    return (
        <Fragment>
            {
                (props.passwordLength > 0 || props.passwordRequireCapital || props.passwordRequireNumeric || props.passwordRequireSpecial) &&
                <h2 style={{justifyContent: "center"}}>Password Requirements</h2>
            }
            {
                props.passwordLength > 0 &&
                <Row style={{justifyContent: "center"}}>Minimum Length of {props.passwordLength} </Row>
            }
            {
                props.passwordRequireCapital &&
                <Row style={{justifyContent: "center"}}>Requires at least 1 Capital Letter</Row>
            }
            {
                props.passwordRequireNumeric &&
                <Row style={{justifyContent: "center"}}>Requires at least 1 Digit</Row>
            }
            {
                props.passwordRequireSpecial &&
                <Fragment>
                    <Row style={{justifyContent: "center"}}>Requires at least 1 Special Character</Row>
                    <Row style={{justifyContent: "center"}}>{props.passwordValidSpecialCharacters.split("").join(" ")}</Row>
                </Fragment>
            }
            <Container>
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