import { useState } from "react";
import * as account from '../../api/account';
import { Button, Container, Form, Row } from "react-bootstrap";
import { Link } from "react-router-dom";

const SendRequest = (props) => {
    const [email, setEmail] = useState('');
    const [complete, setComplete] = useState(false);

    const handleSubmit = async e => {
        e.preventDefault();
        console.log(email);

        account.forgotPassword({ body: { "email": email } })
            .then(resp => {
                if (!resp.error) {
                    setComplete(true)
                } else {
                    setComplete(true)
                }
            })
            .catch((resp) => {
                setComplete(true)
            })
    };

    if (!complete) {
        return (
            <Container fluid>
                <Row>
                    <img alt="ziggy" src={"/ziggy.svg"} width={200} />
                </Row>
                <Row>
                    <h1>zrok</h1>
                </Row>
                <Row>
                    <h2>Forgot Password</h2>
                </Row>
                <Row className={"fullscreen-body"}>
                    <Container className={"fullscreen-form"}>
                        <Row>
                            <Form onSubmit={handleSubmit}>
                                <Form.Group controlId={"email"}>
                                    <Form.Control
                                        type={"email"}
                                        placeholder={"Email Address"}
                                        onChange={t => { setEmail(t.target.value); }}
                                        value={email}
                                    />
                                </Form.Group>

                                <Button variant={"light"} type={"submit"}>Forgot Password</Button>
                            </Form>
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
                    <h1>Reset Password</h1>
                </Row>
                <Row>
                    We will get back to you shortly with a link to reset your password!
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

export default SendRequest;