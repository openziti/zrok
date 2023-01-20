import {useLocation, useParams} from "react-router-dom";
import SendRequest from "./SendRequest"
import SetNewPassword from "./SetNewPassword";

const ResetPassword = () => {
    const { search } = useLocation();
    const { token } = useParams();
    console.log(token)
    let component = undefined
    if (token) {
        component = <SetNewPassword token={token} />
    } else {
        component = <SendRequest />
    }

    console.log(token);

    return (
        <div className={"fullscreen"}>{component}</div>
    )
}

export default ResetPassword;