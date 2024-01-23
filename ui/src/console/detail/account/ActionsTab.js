import React, {useState} from "react";
import ChangePassword from "./actions/ChangePassword";

const ActionsTab = (props) => {
    const [actionState, setActionState] = useState("menu")

    let defaultActionsTabComponent = (
        <div>
            <button onClick={()=>setActionState("changePassword")}>Change Password</button>
        </div>
    )
    const returnState = () => {
        setActionState("menu")
    }

    const renderActions = () => {
        switch (actionState) {
            case "changePassword":
                return <ChangePassword user={props.user} returnToActions={returnState}/>
            default:
                return defaultActionsTabComponent
        }
    }

    return (
        renderActions()
    )
}

export default ActionsTab;