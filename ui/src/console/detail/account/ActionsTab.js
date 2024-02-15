import React, {useState} from "react";
import ChangePassword from "./actions/ChangePassword";
import ResetToken from "./actions/ResetToken";
import {Button} from "react-bootstrap";

const ActionsTab = (props) => {
    const [showResetTokenModal, setShowResetTokenModal] = useState(false);
    const openResetTokenModal = () => setShowResetTokenModal(true);
    const closeResetTokenModal = () => setShowResetTokenModal(false);

    return (
        <div className={"actions-tab"}>
            <div id={"token-regeneration"}>
                <h3>Regenerate your account token <strong>(DANGER!)</strong>?</h3>
                <p>
                    Regenerating your account token will stop all environments and shares from operating properly!
                </p>
                <p>
                    You will need to <strong>manually</strong> edit your
                    <code> &#36;&#123;HOME&#125;/.zrok/environment.json</code> files (in each environment) to use the new
                    <code> zrok_token</code> . Updating these files will restore the functionality of your environments.
                </p>
                <p>
                    Alternatively, you can just <code>zrok disable</code> any enabled environments and re-enable using the
                    new account token. Running <code>zrok disable</code> will <strong>delete</strong> your environments and
                    any shares they contain (including reserved shares). So if you have environments and reserved shares you
                    need to preserve, your best bet is to update the <code>zrok_token</code> in those environments as
                    described above.
                </p>
                <Button variant={"danger"} onClick={openResetTokenModal}>Regenerate Account Token</Button>
                <ResetToken show={showResetTokenModal} onHide={closeResetTokenModal} user={props.user}/>
            </div>
        </div>
    )
}

export default ActionsTab;