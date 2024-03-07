import {useEffect, useState} from "react";
import Modal from "react-bootstrap/Modal";
import { metadataApi } from "../..";

const Version = (props) => {
    const [v, setV] = useState('');

    useEffect(() => {
        let mounted = true;
        metadataApi.version().then(resp => {
            if(mounted) {
                setV(resp);
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