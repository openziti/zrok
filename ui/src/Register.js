import { useParams } from 'react-router-dom';
import {useEffect, useState} from "react";
import * as identity from "./api/identity";

const Proceed = (props) => {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [message, setMessage] = useState();

    const passwordMismatchMessage = <h2 className={"errorMessage"}>Entered passwords do not match!</h2>
    const registerFailed = <h2 className={"errorMessage"}>Account creation failed!</h2>

    const handleSubmit = async e => {
        e.preventDefault();
        console.log("handleSubmit");
        if(confirm !== password) {
            setMessage(passwordMismatchMessage);
        } else {
            identity.register({body: {"token": props.token, "password": password}})
                .then(resp => {
                    if(!resp.error) {
                        console.log("resp", resp)
                        setMessage(undefined);
                    } else {
                        setMessage(registerFailed);
                    }
                })
                .catch(resp => {
                    console.log("resp", resp);
                    setMessage(registerFailed);
                });
        }
    };

    return (
        <div className={"fullscreen"}>
            <img src={"/ziggy.svg"} width={200}/>
            <h1>A new zrok user!</h1>
            <h2>{props.email}</h2>
            <form onSubmit={handleSubmit}>
                <fieldset>
                    <legend>Set A Password</legend>
                    <p><label htmlFor={"password"}>password: </label><input type={"password"} value={password} placeholder={"Password"} onChange={({target}) => setPassword(target.value)}/></p>
                    <p>
                        <label htmlFor={"confirm"}>confirm: </label><input type={"password"} value={confirm} placeholder={"Confirm Password"} onChange={({target}) => setConfirm(target.value)}/>
                        <button type={"submit"}>Register</button>
                    </p>
                </fieldset>
            </form>
            {message}
        </div>
    )
}

const Failed = () => {
    return (
        <div className={"fullscreen"}>
            <img src={"/ziggy.svg"} width={200}/>
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
        step = <Proceed email={email} token={token}/>
    } else {
        step = <Failed />
    }

    return (
        <div>{step}</div>
    )
}

export default Register;