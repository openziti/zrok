import * as share from "../../../api/share";
import {Button} from "react-bootstrap";

const ActionsTab = (props) => {
    const deleteFrontend = (feToken, shrToken, envZId) => {
        if(window.confirm("Really delete access frontend '" + feToken + "' for share '" + shrToken + "'?")) {
            share.unaccess({body: {frontendToken: feToken, shrToken: shrToken, envZId: envZId}}).then(resp => {
                console.log(resp);
            });
        }
    }

    return (
        <div className={"actions-tab"}>
            <h3>Delete your access frontend '{props.frontend.token}' for share '{props.frontend.shrToken}'?</h3>
            <p>
                This will remove your <code>zrok access</code> frontend from this environment. You will still need to
                terminate the corresponding <code>zrok access</code> process in your local environment.
            </p>
            <Button variant={"danger"} onClick={() => deleteFrontend(props.frontend.token, props.frontend.shrToken, props.frontend.zId)}>
                Delete '{props.frontend.token}'
            </Button>
        </div>
    );
};

export default ActionsTab;