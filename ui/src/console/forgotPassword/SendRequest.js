import { useState } from "react";
import * as account from '../../api/account';
import { Button, Container, Form, Row } from "react-bootstrap";

const SendRequest = (props) => {
    const [email, setEmail] = useState('');
    const [message, setMessage] = useState();
    const [complete, setComplete] = useState(false);


    const errorMessage = <h2 className={"errorMessage"}>Forgot Password Failed!</h2>;

    const handleSubmit = async e => {
        e.preventDefault();
        console.log(email);

        account.forgotPassword({ body: { "email": email } })
            .then(resp => {
                if (!resp.error) {
                    console.log("Make landing page to expect and email or something similar")
                    setComplete(true)
                } else {
                    console.log('forgot password failed')
                    setMessage(errorMessage);
                }
            })
            .catch((resp) => {
                console.log('forgot password failed', resp)
                setMessage(errorMessage)
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
                                        onChange={t => { setMessage(null); setEmail(t.target.value); }}
                                        value={email}
                                    />
                                </Form.Group>

                                <Button variant={"light"} type={"submit"}>Forgot Password</Button>
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
        <div>Make landing page to expect an email or something similar</div>
    )
}

export default SendRequest;