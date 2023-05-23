import {useEffect, useState} from "react";
import * as account from '../api/account';
import * as metadata from "../api/metadata"
import {Button, Container, Form, Row} from "react-bootstrap";
import { Link } from "react-router-dom";
import PasswordForm from "../components/password";

const SetNewPassword = (props) => {
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState();
    const [complete, setComplete] = useState(false);
    const [passwordLength, setPasswordLength] = useState(10);
    const [passwordRequireCapital, setPasswordRequireCapital] = useState(true);
    const [passwordRequireNumeric, setPasswordRequireNumeric] = useState(true);
    const [passwordRequireSpecial, setPasswordRequireSpecial] = useState(true);
    const [passwordValidSpecialCharacters, setPasswordValidSpecialCharacters] = useState("");

    const errorMessage = <h2 className={"errorMessage"}>Reset Password Failed!</h2>;

    useEffect(() => {
        metadata.configuration().then(resp => {
            if(!resp.error) {
                setPasswordLength(resp.data.passwordRequirements.length)
                setPasswordRequireCapital(resp.data.passwordRequirements.requireCapital)
                setPasswordRequireNumeric(resp.data.passwordRequirements.requireNumeric)
                setPasswordRequireSpecial(resp.data.passwordRequirements.requireSpecial)
                setPasswordValidSpecialCharacters(resp.data.passwordRequirements.validSpecialCharacters)
            }
        }).catch(err => {
            console.log("err", err);
        });
    }, [])

    const handleSubmit = async e => {
        e.preventDefault();
        if (password !== undefined && password !== "") {
        account.resetPassword({body: {"token": props.token, "password": password}})
            .then(resp => {
                if(!resp.error) {
                    setMessage(undefined);
                    setComplete(true);
                } else {
                    setMessage(errorMessage);
                }
            })
            .catch(resp => {
                setMessage(errorMessage);
            })
        }
    }

    if(!complete) {
        return (
            <Container fluid>
                <Row>
                    <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
                </Row>
                <Row>
                    <h1>Reset Password</h1>
                </Row>
                <Row className={"fullscreen-body"}>
                    <Container className={"fullscreen-form"}>
                        <Row>
                            <Form onSubmit={handleSubmit}>
                            <PasswordForm
                                    setMessage={setMessage}
                                    passwordLength={passwordLength}
                                    passwordRequireCapital={passwordRequireCapital}
                                    passwordRequireNumeric={passwordRequireNumeric}
                                    passwordRequireSpecial={passwordRequireSpecial}
                                    passwordValidSpecialCharacters={passwordValidSpecialCharacters}
                                    setParentPassword={setPassword}/>

                                <Button variant={"light"} type={"submit"}>Reset Password</Button>
                            </Form>
                        </Row>
                        <Row>
                            {message}
                        </Row>
                    </Container>
                </Row>
            </Container>
        )
    }

    return (
        <Container fluid>
            <Row>
                <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
            </Row>
            <Row>
                <h1>Password Reset</h1>
            </Row>
            <Row>
                Password reset successful! You can now return to the login page and login.
            </Row>
            <Row>
                <div id={"zrok-reset-password"}>
                    <Link to="/" className="">
                        <Button variant={"light"}>Login</Button>
                    </Link>
                </div>
            </Row>
        </Container>
    )
}

export default SetNewPassword;
