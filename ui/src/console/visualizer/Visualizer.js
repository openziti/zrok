import React, {useEffect, useState} from "react";
import {Button} from "react-bootstrap";
import Network from "./Network";
import {mergeGraph} from "./graph";
import {isSelectionGone, markSelected} from "./selection";
import {mdiFitToPageOutline} from "@mdi/js";
import Icon from "@mdi/react";

const Visualizer = (props) => {
    const [networkGraph, setNetworkGraph] = useState({nodes: [], links: []});

    useEffect(() => {
        setNetworkGraph(mergeGraph(networkGraph, props.user, props.overview.accountLimited, props.overview.environments));

        if(isSelectionGone(networkGraph, props.selection)) {
            // if the selection is no longer in the network graph...
            console.log("resetting selection", props.selection);
            props.setSelection(props.defaultSelection);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [props.overview]);

    markSelected(networkGraph, props.selection);

    useEffect(() => {
        markSelected(networkGraph, props.selection);
        // eslint-disable-next-line react-hooks/exhaustive-deps
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
                    <Button variant={"secondary"} size={"sm"} onClick={centerFocus} aria-label={"Zoom to Fit"}><Icon path={mdiFitToPageOutline} size={0.7}/></Button>
                </div>
            </div>
        </div>
    )
}

export default Visualizer;