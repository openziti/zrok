import React, {useState} from "react";
import ChangePassword from "./actions/ChangePassword";
import {Button} from "react-bootstrap";
import RegenerateToken from "./actions/RegenerateToken";

const ActionsTab = (props) => {
    const [showRegenerateTokenModal, setShowRegenerateTokenModal] = useState(false);
    const openRegenerateTokenModal = () => setShowRegenerateTokenModal(true);
    const closeRegenerateTokenModal = () => setShowRegenerateTokenModal(false);

    const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
    const openChangePasswordModal = () => setShowChangePasswordModal(true);
    const closeChangePasswordModal = () => setShowChangePasswordModal(false);

    return (
        <div className={"actions-tab"}>
            <div id={"change-password"} style={{"paddingTop": "10px"}}>
                <h3>Change Password?</h3>
                <p>Change the password used to log into the zrok web console.</p>
                <Button variant={"danger"} onClick={openChangePasswordModal}>Change Password</Button>
                <ChangePassword show={showChangePasswordModal} onHide={closeChangePasswordModal} user={props.user}/>
            </div>

            <hr/>

            <div id={"token-regeneration"}>
                <h3>DANGER: Regenerate your account token?</h3>
                <p>
                    Regenerating your account token will stop all environments and shares from operating properly!
                </p>
                <p>
                    You will need to <em><strong>manually</strong></em> edit your
                    <code> &#36;&#123;HOME&#125;/.zrok/environment.json</code> files (in each environment) to use the new
                    <code> zrok_token</code>. Updating these files will restore the functionality of your environments.
                </p>
                <p>
                    Alternatively, you can just <code>zrok disable</code> any enabled environments and re-enable using the
                    new account token. Running <code>zrok disable</code> will <em><strong>delete</strong></em> your environments and
                    any shares they contain (including reserved shares). So if you have environments and reserved shares you
                    need to preserve, your best option is to update the <code>zrok_token</code> in those environments as
                    described above.
                </p>
                <Button variant={"danger"} onClick={openRegenerateTokenModal}>Regenerate Account Token</Button>
                <RegenerateToken show={showRegenerateTokenModal} onHide={closeRegenerateTokenModal} user={props.user}/>
            </div>
            </div>
    )
}

export default ActionsTab;