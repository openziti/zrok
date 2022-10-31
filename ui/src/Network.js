import Environments from './Environments';
import ReactFlow from "react-flow-renderer";

const Network = (props) => {
    return (
        <div>
            <div className={"network"}>
                <h1>Network</h1>
                <ReactFlow
                    nodes={props.nodes}
                    edges={props.edges}
                />
            </div>
            <Environments
                overview={props.overview}
            />
        </div>
    )
}

export default Network;