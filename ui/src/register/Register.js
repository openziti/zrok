import { useParams } from 'react-router-dom';
import {useEffect, useState} from "react";
import * as account from "../api/account";
import InvalidRequest from "./InvalidRequest";
import SetPasswordForm from "./SetPasswordForm";

let step;

const Register = () => {
    const { token } = useParams();
    const [email, setEmail] = useState();
    const [activeRequest, setActiveRequest] = useState(true);

    useEffect(() => {
        let mounted = true
        account.verify({body: {token: token}}).then(resp => {
            if(mounted) {
                if(resp.error) {
                    setActiveRequest(false);
                } else {
                    setEmail(resp.data.email);
                }
            }
        }).catch(err => {
            console.log("err", err);
            if(mounted) {
                setActiveRequest(false);
            }
        });
        return () => {
            mounted = false;
        }
    }, [token]);

    if(activeRequest) {
        step = <SetPasswordForm email={email} token={token}/>
    } else {
        step = <InvalidRequest />
    }

    return (
        <div className={"fullscreen"}>{step}</div>
    )
}

export default Register;