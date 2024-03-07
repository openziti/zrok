import { useParams } from 'react-router-dom';
import {useEffect, useState} from "react";
import InvalidRequest from "./InvalidRequest";
import SetPasswordForm from "./SetPasswordForm";
import { accountApi } from '..';

let step;

const Register = () => {
    const { token } = useParams();
    const [email, setEmail] = useState();
    const [activeRequest, setActiveRequest] = useState(true);

    useEffect(() => {
        let mounted = true
        accountApi.verify({body: {token: token}}).then(resp => {
            if(mounted) {
                setEmail(resp.email);
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