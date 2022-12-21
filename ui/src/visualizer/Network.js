import {withSize} from "react-sizeme";
import {useRef} from "react";

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
            backgroundColor={"#30205d"}
        />
    )
}

export default withSize()(Network);