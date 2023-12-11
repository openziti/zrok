import {Button} from "react-bootstrap";
import {EnvironmentApi} from "../../../api/src";

const ActionsTab = (props) => {

    const environment = new EnvironmentApi()

    const deleteEnvironment = (envZId) => {
        if(window.confirm("Really delete environment '" + envZId + "' and all shares within?")) {
            environment.disable({body: {identity: envZId}}).then(resp => {
                console.log(resp);
            });
        }
    }

    return (
        <div className={"actions-tab"}>
            <h3>Delete your environment '{props.environment.description}' ({props.environment.zId})?</h3>
            <p>
                This will remove all shares from this environment, and will remove the environment from the network. You
                will still need to terminate backends and <code>disable</code> your local environment.
            </p>
            <Button variant={"danger"} onClick={() => deleteEnvironment(props.environment.zId)}>Delete '{props.environment.description}'</Button>
        </div>
    );
};

export default ActionsTab;