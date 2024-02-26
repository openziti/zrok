import { useState } from "react";
import { Button, Container, Form, Row } from "react-bootstrap";
import { accountApi } from "..";

const SendRequest = () => {
    const [email, setEmail] = useState('');
    const [complete, setComplete] = useState(false);

    const handleSubmit = async e => {
        e.preventDefault();
        console.log(email);

        accountApi.resetPasswordRequest({ body: { "emailAddress": email } })
            .then(resp => {
                setComplete(true)
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
                    Check your email for a password reset message!
                </Row>
        </Container>
    )
}

export default SendRequest;