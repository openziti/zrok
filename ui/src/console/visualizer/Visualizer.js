import React, {useEffect, useState} from "react";
import {Button} from "react-bootstrap";
import Network from "./Network";
import {mergeGraph} from "./graph";
import {isSelectionGone, markSelected} from "./selection";

const Visualizer = (props) => {
    const [networkGraph, setNetworkGraph] = useState({nodes: [], links: []});

    useEffect(() => {
        setNetworkGraph(mergeGraph(networkGraph, props.user, props.overview));

        if(isSelectionGone(networkGraph, props.selection)) {
            // if the selection is no longer in the network graph...
            console.log("resetting selection", props.selection);
            props.setSelection(props.defaultSelection);
        }
    }, [props.overview]);

    markSelected(networkGraph, props.selection);

    useEffect(() => {
        markSelected(networkGraph, props.selection);
    }, [props.selection]);

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
                    setSelection={props.setSelection}
                />
                <div className={"visualizer-controls"}>
                    <Button variant={"secondary"} size={"sm"} onClick={centerFocus}>Zoom to Fit</Button>
                </div>
            </div>
        </div>
    )
}

export default Visualizer;