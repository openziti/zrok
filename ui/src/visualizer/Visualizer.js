import React, {useState} from "react";
import {Button, Container, Row} from "react-bootstrap";
import Network from "./Network";

const Visualizer = (props) => {
    const [networkGraph, setNetworkGraph] = useState({nodes: [{id: 1}], links: []});
    const [fgRef, setFgRef] = useState(() => {});

    const centerFocus = () => {
        if(fgRef != null) {
            fgRef.current.zoomToFit(200);
        }
    }

    return (
        <div>
            <div className={"visualizer-container"}>
                <Network
                    networkGraph={networkGraph}
                    setRef={setFgRef}
                />
            </div>
            <div className={"visualizer-controls"}>
                <Button onClick={centerFocus}>Zoom to Fit</Button>
            </div>
        </div>
    )
}

export default Visualizer;