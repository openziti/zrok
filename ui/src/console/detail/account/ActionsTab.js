import React, {useState} from "react";
import ResetToken from "./actions/ResetToken";

const ActionsTab = (props) => {
    const [actionState, setActionState] = useState("menu")

    const [showResetTokenModal, setShowResetTokenModal] = useState(false);
    const openResetTokenModal = () => setShowResetTokenModal(true);
    const closeResetTokenModal = () => setShowResetTokenModal(false);

    let defaultActionsTabComponent = (
        <div>
            <button onClick={openResetTokenModal}>Regenerate Token</button>
            <ResetToken show={showResetTokenModal} onHide={closeResetTokenModal} user={props.user}/>
        </div>
    )

    const renderActions = () => {
        switch (actionState) {
            default:
                return defaultActionsTabComponent
        }
    }

    return (
        renderActions()
    )
}

export default ActionsTab;

