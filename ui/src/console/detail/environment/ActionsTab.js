import {Button} from "react-bootstrap";

const ActionsTab = (props) => {
    return (
        <div class={"actions-tab"}>
            <h3>Delete your environment '{props.environment.description}' ({props.environment.zId})?</h3>
            <p>
                This will remove all shares from this environment, and will remove the environment from the network. You
                will still need to terminate backends and <code>disable</code> your local environment.
            </p>
            <Button variant={"danger"}>Delete '{props.environment.description}'</Button>
        </div>
    );
};

export default ActionsTab;