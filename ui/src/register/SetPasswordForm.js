import React, {useEffect, useState} from "react";
import * as account from "../api/account";
import * as metadata from "../api/metadata"
import Success from "./Success";
import {Button, Container, Form, Row} from "react-bootstrap";

const SetPasswordForm = (props) => {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [message, setMessage] = useState();
    const [authToken, setAuthToken] = useState('');
    const [complete, setComplete] = useState(false);
    const [tou, setTou] = useState();

    const passwordMismatchMessage = <h2 className={"errorMessage"}>Entered passwords do not match!</h2>
    const passwordTooShortMessage = <h2 className={"errorMessage"}>Entered password too short! (4 characters, minimum)</h2>
    const registerFailed = <h2 className={"errorMessage"}>Account creation failed!</h2>

    useEffect(() => {
        metadata.configuration().then(resp => {
            console.log(resp)
            if(!resp.error) {
                if (resp.data.touLink !== null && resp.data.touLink.trim() !== "") {
                    setTou(resp.data.touLink)
                }
            }
        }).catch(err => {
            console.log("err", err);
        });
    }, [])

    const handleSubmit = async e => {
        e.preventDefault();
        if(confirm.length < 4) {
            setMessage(passwordTooShortMessage);
            return;
        }
        if(confirm !== password) {
            setMessage(passwordMismatchMessage);
            return;
        }
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
                                <Form.Group controlId={"password"}>
                                    <Form.Control
                                        type={"password"}
                                        placeholder={"Set Password"}
                                        onChange={t => { setMessage(null); setPassword(t.target.value); }}
                                        value={password}
                                    />
                                </Form.Group>

                                <Form.Group controlId={"confirm"}>
                                    <Form.Control
                                        type={"password"}
                                        placeholder={"Confirm Password"}
                                        onChange={t => { setMessage(null); setConfirm(t.target.value); }}
                                        value={confirm}
                                    />
                                </Form.Group>
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