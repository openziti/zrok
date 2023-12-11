import React, {useEffect, useState} from "react";
import {MetadataApi, AccountApi} from "../api/src"
import Success from "./Success";
import {Button, Container, Form, Row} from "react-bootstrap";
import PasswordForm from "../components/password";

const SetPasswordForm = (props) => {
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState();
    const [authToken, setAuthToken] = useState('');
    const [complete, setComplete] = useState(false);
    const [tou, setTou] = useState();
    const [passwordLength, setPasswordLength] = useState(10);
    const [passwordRequireCapital, setPasswordRequireCapital] = useState(true);
    const [passwordRequireNumeric, setPasswordRequireNumeric] = useState(true);
    const [passwordRequireSpecial, setPasswordRequireSpecial] = useState(true);
    const [passwordValidSpecialCharacters, setPasswordValidSpecialCharacters] = useState("");

    const registerFailed = <h2 className={"errorMessage"}>Account creation failed!</h2>

    const metadata = new MetadataApi()
    const account = new AccountApi()

    useEffect(() => {
        metadata.configuration().then(resp => {
            if(!resp.error) {
                if (resp.touLink !== undefined && resp.touLink.trim() !== "") {
                    setTou(resp.touLink)
                }
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
        account.register({body: {"token": props.token, "password": password}})
            .then(resp => {
                console.log("resp", resp)
                setMessage(undefined);
                setAuthToken(resp.token);
                setComplete(true);
            })
            .catch(resp => {
                console.log("resp", resp);
                setMessage(registerFailed);
            });
        }
    };

    if(!complete) {
        return (
            <Container fluid>
                <Row>
                    <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
                </Row>
                <Row>
                    <h1>Welcome new zrok user!</h1>
                </Row>
                <Row>
                    <h2>{props.email}</h2>
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
                                <Button variant={"light"} type={"submit"}>Register Account</Button>
                            </Form>
                        </Row>
                        <Row>
                            {message}
                        </Row>
                        <Row>
                            <div id={"zrok-tou"} dangerouslySetInnerHTML={{__html: tou}}></div>
                        </Row>
                    </Container>
                </Row>
            </Container>
        );
    }
    return <Success email={props.email} token={authToken}/>;
};

export default SetPasswordForm;