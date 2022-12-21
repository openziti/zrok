import {useEffect, useState} from "react";
import * as metadata from "../../api/metadata";
import Modal from "react-bootstrap/Modal";

const Version = (props) => {
    const [v, setV] = useState('');

    useEffect(() => {
        let mounted = true;
        metadata.version().then(resp => {
            if(mounted) {
                setV(resp.data);
            }
        });
        return () => {
            mounted = false;
        };
    }, []);

    return (
        <Modal show={props.show} onHide={props.onHide} centered>
            <Modal.Header closeButton>About zrok</Modal.Header>
            <Modal.Body>
                <code>{v}</code>
            </Modal.Body>
        </Modal>
    );
}

export default Version;