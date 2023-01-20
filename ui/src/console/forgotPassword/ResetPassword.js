import {useState} from "react";
import * as account from '../../api/account';
import {Button, Container, Form, Row} from "react-bootstrap";
import { Link } from "react-router-dom";

const ResetPassword = (props) => {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [message, setMessage] = useState();
    const [complete, setComplete] = useState(false);

    const passwordMismatchMessage = <h2 className={"errorMessage"}>Entered passwords do not match!</h2>
    const passwordTooShortMessage = <h2 className={"errorMessage"}>Entered password too short! (4 characters, minimum)</h2>

    const errorMessage = <h2 className={"errorMessage"}>Reset Password Failed!</h2>;

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
                <div>
                    <Link to="/" className="">
                        Login
                    </Link>
                </div>
            </Row>
        </Container>
    )
}

export default ResetPassword;
