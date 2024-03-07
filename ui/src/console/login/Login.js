import {useEffect, useState} from "react";
import {Button, Container, Form, Row} from "react-bootstrap";
import { Link } from "react-router-dom";
import { accountApi } from "../..";
import { metadataApi } from "../..";

const Login = (props) => {
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState();
    const [tou, setTou] = useState();

    const errorMessage = <h2 className={"errorMessage"}>Login Failed!</h2>;

    useEffect(() => {
        metadataApi.configuration().then(resp => {
            console.log(resp)
            if (resp.touLink !== null && resp.touLink.trim() !== "") {
                setTou(resp.touLink)
            }
        }).catch(err => {
            console.log("err", err);
        });
    }, [])

    const handleSubmit = async e => {
        e.preventDefault()
        console.log(email, password);

        accountApi.login({body: {"email": email, "password": password}})
            .then(resp => {
                let user = {
                    "email": email.toLowerCase(),
                    "token": resp
                }
                props.loginSuccess(user)
                localStorage.setItem('user', JSON.stringify(user))
                console.log(user)
                console.log('login succeeded', resp)
                document.dispatchEvent(new Event('storage'))
            })
            .catch((resp) => {
                console.log('login failed', resp)
                setMessage(errorMessage);
            });
    };

    return (
        <div className={"fullscreen"}>
            <Container fluid>
                <Row>
                    <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
                </Row>
                <Row>
                    <h1>zrok</h1>
                </Row>
                <Row className={"fullscreen-body"}>
                    <Container className={"fullscreen-form"}>
                        <Row>
                            <Form onSubmit={handleSubmit}>
                                <Form.Group controlId={"email"}>
                                    <Form.Control
                                        type={"email"}
                                        placeholder={"Email Address"}
                                        onChange={t => { setMessage(null); setEmail(t.target.value); }}
                                        value={email}
                                    />
                                </Form.Group>

                                <Form.Group controlId={"password"}>
                                    <Form.Control
                                        type={"password"}
                                        placeholder={"Password"}
                                        onChange={t => { setMessage(null); setPassword(t.target.value); }}
                                        value={password}
                                    />
                                </Form.Group>

                                <Button variant={"light"} type={"submit"}>Log In</Button>
                                
                                <div id={"zrok-reset-password"}>
                                    <Link to="/resetPassword" className="">
                                        Forgot Password?
                                    </Link>
                                </div>
                                <div id={"zrok-tou"} dangerouslySetInnerHTML={{__html: tou}}></div>
                            </Form>
                        </Row>
                        <Row>
                            {message}
                        </Row>
                    </Container>
                </Row>
            </Container>
        </div>
    )
}

export default Login;