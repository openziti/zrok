import Modal from "react-bootstrap/Modal";
import { Button, Container, Form, Row } from "react-bootstrap";
import * as account from "../../../../api/account";

const ResetToken = (props) => {

    let resetToken = () => {
        console.log("I should reset my token")
        account.resetToken({ body: { "emailAddress": props.user.email } }).then(resp => {
            console.log(resp)
        }).catch(err => {
            console.log("err", err);
        });
    }

    return (
        <div>
            <Modal show={props.show} onHide={props.onHide} centered>
                <Modal.Header closeButton>Are you Sure?</Modal.Header>
                <Modal.Body>
                    TEST
                    <Button variant={"light"} onClick={resetToken}>Reset Password</Button>
                </Modal.Body>
            </Modal>
        </div>
    )
}

export default ResetToken;