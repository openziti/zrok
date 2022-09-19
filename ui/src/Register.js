import { useParams } from 'react-router-dom';
import {useEffect, useState} from "react";
import * as identity from "./api/identity";

const Proceed = (props) => {
    return (
        <div>
            <h1>Register a new zrok account!</h1>
            <h3>{props.email}</h3>
        </div>
    )
}

const Failed = () => {
    return (
        <div>
            <h1>No such account request!</h1>
        </div>
    )
}

const Register = () => {
    const { token } = useParams();
    const [email, setEmail] = useState();
    const [failed, setFailed] = useState(false);

    useEffect(() => {
        let mounted = true
        identity.verify({body: {token: token}}).then(resp => {
            console.log("resp", resp)
            if(mounted) {
                if(resp.error) {
                    setFailed(true);
                } else {
                    setEmail(resp.data.email);
                }
            }
        }).catch(err => {
            console.log("err", err);
            if(mounted) {
                setFailed(true);
            }
        });
        return () => {
            mounted = false;
        }
    }, []);

    let step;
    if(!failed) {
        step = <Proceed email={email}/>
    } else {
        step = <Failed />
    }

    return (
        <div className={"zrok"}>
            <div className={"container"}>
                <div className={"header"}>
                    <img alt={"ziggy goes to space"} src="/ziggy.svg" width={"65px"} />
                    <p className={"header-title"}>zrok</p>
                </div>
                <div className={"main"}>
                    {step}
                </div>
            </div>
        </div>
    )
}

export default Register;