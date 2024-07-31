import React, {useEffect, useState} from "react";
import * as account from "../api/account";
import * as metadata from "../api/metadata"
import Success from "./Success";
import {Button, Container, Form, Row} from "react-bootstrap";
import PasswordForm from "../components/password";
import {Checkbox, FormControl, FormControlLabel, FormGroup, InputLabel} from "@mui/material";

const SetPasswordForm = (props) => {
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState();
    const [authToken, setAuthToken] = useState('');
    const [complete, setComplete] = useState(false);
    const [tou, setTou] = useState(<span/>);
    const [passwordLength, setPasswordLength] = useState(10);
    const [passwordRequireCapital, setPasswordRequireCapital] = useState(true);
    const [passwordRequireNumeric, setPasswordRequireNumeric] = useState(true);
    const [passwordRequireSpecial, setPasswordRequireSpecial] = useState(true);
    const [passwordValidSpecialCharacters, setPasswordValidSpecialCharacters] = useState("");

    const registerFailed = <h2 className={"errorMessage"}>Account creation failed!</h2>

    useEffect(() => {
        metadata.configuration().then(resp => {
            if(!resp.error) {
                if (resp.data.touLink !== undefined && resp.data.touLink.trim() !== "") {
                    setTou(<span dangerouslySetInnerHTML={{__html: resp.data.touLink}}/>)
                }
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
        account.register({body: {"token": props.token, "password": password}})
            .then(resp => {
                if(!resp.error) {
                    console.log("resp", resp)
                    setMessage(undefined);
                    setAuthToken(resp.data.token);
                    setComplete(true);
                } else {
                    setMessage(registerFailed);
                }
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
                                <div>
                                    <PasswordForm
                                        setMessage={setMessage}
                                        passwordLength={passwordLength}
                                        passwordRequireCapital={passwordRequireCapital}
                                        passwordRequireNumeric={passwordRequireNumeric}
                                        passwordRequireSpecial={passwordRequireSpecial}
                                        passwordValidSpecialCharacters={passwordValidSpecialCharacters}
                                        setParentPassword={setPassword}/>
                                </div>
                                <div class={"zrok-tou"}>
                                    <p>{tou}</p>
                                </div>
                                <div class={"zrok-tou"}>
                                    <FormControlLabel control={<Checkbox style={{color: 'white'}} required/>} label={<span class={"zrok-type"}>I have read and agree to the above</span>}/>
                                </div>
                                <Button variant={"light"} type={"submit"}>Register Account</Button>
                            </Form>
                        </Row>
                        <Row id={"zrok-message-row"}>
                            {message}
                        </Row>
                    </Container>
                </Row>
            </Container>
        );
    }
    return <Success email={props.email} token={authToken}/>;
};

export default SetPasswordForm;