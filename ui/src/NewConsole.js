import {Container, Nav, Navbar, NavDropdown, Row} from "react-bootstrap";
import {useState} from "react";
import Visualizer from "./visualizer/Visualizer";
import NewEnable from "./modals/NewEnable";
import NewVersion from "./modals/NewVersion";

const NewConsole = (props) => {
    const [showEnableModal, setShowEnableModal] = useState(false);
    const openEnableModal = () => setShowEnableModal(true);
    const closeEnableModal = () => setShowEnableModal(false);

    const [showVersionModal, setShowVersionModal] = useState(false);
    const openVersionModal = () => setShowVersionModal(true);
    const closeVersionModal = () => setShowVersionModal(false);

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
            <Visualizer />
            <NewEnable show={showEnableModal} onHide={closeEnableModal} token={props.user.token}/>
            <NewVersion show={showVersionModal} onHide={closeVersionModal} />
        </Container>
    );
}

export default NewConsole;