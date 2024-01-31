import Modal from "react-bootstrap/Modal";
import { Button, Container, Form, Row } from "react-bootstrap";
import * as account from "../../../../api/account";

const ResetToken = (props) => {

    let resetToken = () => {
        console.log("I should reset my token")
        account.resetToken({ body: { "emailAddress": props.user.email } }).then(resp => {
            console.log(resp)
            let user = JSON.parse(localStorage.getItem('user'))
            localStorage.setItem('user', JSON.stringify({
                "email": user.email,
                "token": resp.data.token
            }));
            document.dispatchEvent(new Event('storage'))
        }).catch(err => {
            console.log("err", err);
        });
        props.onHide();
    }

    return (
        <div>
            <Modal show={props.show} onHide={props.onHide} centered>
                <Modal.Header closeButton>WARNING - Are you Sure?</Modal.Header>
                <Modal.Body>
                    <div>
                        Reseting your token will remove all environments, frontends, and shares you've created.
                    </div>
                    <div style={{display: 'flex', alignItems:'center', justifyContent: 'center'}}>
                        <Button variant={"light"} onClick={resetToken}>Reset Password</Button>
                        <Button variant={"dark"} onClick={props.onHide}>Cancel</Button>
                    </div>
                </Modal.Body>
            </Modal>
        </div>
    )
}

export default ResetToken;