import {ShareApi} from "../../../api/src";
import {Button} from "react-bootstrap";

const ActionsTab = (props) => {
    const share = new ShareApi()
    const deleteShare = (envZId, shrToken, reserved) => {
        console.log(envZId, shrToken, reserved);
        if(window.confirm("Really delete share '" + shrToken + "'?")) {
            share.unshare({body: {envZId: envZId, shrToken: shrToken, reserved: reserved}}).then(resp => {
                console.log(resp);
            });
        }
    }

    return (
        <div className={"actions-tab"}>
            <h3>Delete your share '{props.share.token}'?</h3>
            <p>
                This will remove the share (of <code>{props.share.backendProxyEndpoint}</code>) from your environment, making it
                unavailable. You will need to terminate the backend for this share in your local environment.
            </p>
            <Button variant={"danger"} onClick={() => deleteShare(props.share.envZId, props.share.token, props.share.reserved)}>
                Delete '{props.share.token}'
            </Button>
        </div>
    );
};

export default ActionsTab;