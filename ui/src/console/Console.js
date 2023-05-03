import {Col, Container, Nav, Navbar, NavDropdown, Row} from "react-bootstrap";
import {useEffect, useState} from "react";
import Visualizer from "./visualizer/Visualizer";
import Enable from "./modals/Enable";
import Version from "./modals/Version";
import * as metadata from "../api/metadata";
import Detail from "./detail/Detail";

const Console = (props) => {
    const [showEnableModal, setShowEnableModal] = useState(false);
    const openEnableModal = () => setShowEnableModal(true);
    const closeEnableModal = () => setShowEnableModal(false);

    const [showVersionModal, setShowVersionModal] = useState(false);
    const openVersionModal = () => setShowVersionModal(true);
    const closeVersionModal = () => setShowVersionModal(false);

    const [overview, setOverview] = useState([]);

    useEffect(() => {
        let mounted = true;
        metadata.overview().then(resp => {
            if(mounted) {
                setOverview(resp.data);
            }
        });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            metadata.overview().then(resp => {
                if(mounted) {
                    setOverview(resp.data);
                }
            })
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    }, []);

    const defaultSelection = {id: props.user.token, type: "account"};
    const [selection, setSelection] = useState(defaultSelection);

    return (
        <Container fluid={"xl"}>
            <Navbar bg="primary" variant="dark" id="navbar" expand="md">
                <Container fluid>
                    <Navbar.Brand>
                        <img alt="Ziggy" src="/ziggy.svg" width="65" className="d-inline-block align-top" />{' '}
                        <span className="header-title">zrok</span>
                    </Navbar.Brand>
                    <Navbar.Toggle aria-controls="navbarScroll" />
                    <Navbar.Collapse className="justify-content-end">
                        <Nav navbarScroll>
                            <NavDropdown title={props.user.email}>
                                <NavDropdown.Item onClick={openEnableModal}>Enable Your Environment</NavDropdown.Item>
                                <NavDropdown.Item onClick={openVersionModal}>About zrok</NavDropdown.Item>
                                <NavDropdown.Item onClick={props.logout}>Log Out</NavDropdown.Item>
                            </NavDropdown>
                        </Nav>
                    </Navbar.Collapse>
                </Container>
            </Navbar>
            <Container fluid={"xl"}>
                <Row id={"controls-row"}>
                    <Col lg={6}>
                        <Visualizer
                            user={props.user}
                            overview={overview}
                            defaultSelection={defaultSelection}
                            selection={selection}
                            setSelection={setSelection}
                        />
                    </Col>
                    <Col lg={6}>
                        <Detail user={props.user} selection={selection} />
                    </Col>
                </Row>
            </Container>
            <Enable show={showEnableModal} onHide={closeEnableModal} token={props.user.token} />
            <Version show={showVersionModal} onHide={closeVersionModal} />
        </Container>
    );
}

export default Console;