import {withSize} from "react-sizeme";
import {useRef} from "react";
import {ForceGraph2D} from "react-force-graph";

const Network = (props) => {
    const targetRef = useRef();
    if(props.setRef != null) {
        props.setRef(targetRef);
    }

    return (
        <ForceGraph2D
            ref={targetRef}
            graphData={props.networkGraph}
            width={props.size.width}
            height={500}
            linkOpacity={.75}
            backgroundColor={"#3b2693"}
        />
    )
}

export default withSize()(Network);