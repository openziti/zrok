import React, {useEffect, useState} from "react";
import {Button} from "react-bootstrap";
import Network from "./Network";
import {mergeGraph} from "./graph";

const Visualizer = (props) => {
    const [networkGraph, setNetworkGraph] = useState({nodes: [], links: []});

    useEffect(() => {
        setNetworkGraph(mergeGraph(networkGraph, props.overview));
    }, [props.overview]);

    // fgRef to access force graph controls from this component
    let fgRef = () => { };
    const setFgRef = (ref) => { fgRef = ref };

    const centerFocus = () => {
        if(fgRef) {
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
                <div className={"visualizer-controls"}>
                    <Button variant={"secondary"} size={"sm"} onClick={centerFocus}>Zoom to Fit</Button>
                </div>
            </div>
        </div>
    )
}

export default Visualizer;