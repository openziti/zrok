import {useEffect, useState} from "react";
import {MetadataApi, AccountApi, ResetPasswordRequest} from "../api/src"
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

    const metadata = new MetadataApi()
    const account = new AccountApi()

    useEffect(() => {
        metadata.configuration().then(resp => {
            if(!resp.error) {
                setPasswordLength(resp.passwordRequirements.length)
                setPasswordRequireCapital(resp.passwordRequirements.requireCapital)
                setPasswordRequireNumeric(resp.passwordRequirements.requireNumeric)
                setPasswordRequireSpecial(resp.passwordRequirements.requireSpecial)
                setPasswordValidSpecialCharacters(resp.passwordRequirements.validSpecialCharacters)
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
                console.log(resp)
                setMessage(undefined);
                setComplete(true);
            })
            .catch(resp => {
                console.log(resp)
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
