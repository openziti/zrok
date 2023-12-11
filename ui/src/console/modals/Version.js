import {useEffect, useState} from "react";
import {MetadataApi} from "../../api/src";
import Modal from "react-bootstrap/Modal";

const Version = (props) => {
    const [v, setV] = useState('');

    const metadata = new MetadataApi()

    useEffect(() => {
        let mounted = true;
        metadata.version().then(resp => {
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