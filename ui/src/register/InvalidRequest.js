import {Button, Container, Row} from "react-bootstrap";

const InvalidRequest = () => {
    return (
         <Container fluid>
             <Row>
                 <img alt="ziggy" src={"/ziggy.svg"} width={200}/>
             </Row>
             <Row>
                 <h1>Invitation not found!</h1>
             </Row>
         </Container>
    );
};

export default InvalidRequest;