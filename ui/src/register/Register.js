import Icon from "@mdi/react";
import { useParams } from 'react-router-dom';
import {useEffect, useState} from "react";
import {mdiContentCopy} from "@mdi/js";
import * as account from "../api/account";

const RegistrationForm = (props) => {
    const [password, setPassword] = useState('');
    const [confirm, setConfirm] = useState('');
    const [message, setMessage] = useState();
    const [authToken, setAuthToken] = useState('');
    const [complete, setComplete] = useState(false);

    const passwordMismatchMessage = <h2 className={"errorMessage"}>Entered passwords do not match!</h2>
    const registerFailed = <h2 className={"errorMessage"}>Account creation failed!</h2>

    const handleSubmit = async e => {
        e.preventDefault();
        console.log("handleSubmit");
        if(confirm !== password) {
            setMessage(passwordMismatchMessage);
        } else {
            account.register({body: {"token": props.token, "password": password}})
                .then(resp => {
                    if(!resp.error) {
                        console.log("resp", resp)
                        setMessage(undefined);
                        setAuthToken(resp.data.token);
                        setComplete(true);
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

    if(!complete) {
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
    } else {
        return <Success email={props.email} token={authToken}/>
    }
}

const NoAccountRequest = () => {
    return (
        <div className={"fullscreen"}>
            <img src={"/ziggy.svg"} width={200}/>
            <h1>No such account request!</h1>
        </div>
    )
}

const Success = (props) => {
    const handleCopy = async () => {
        let copiedText = document.getElementById("zrok-enable-command").innerHTML;
        try {
            await navigator.clipboard.writeText(copiedText);
            console.log("copied enable command");
        } catch(err) {
            console.error("failed to copy", err);
        }
    }

    // clear local storage on new account registration success.
    localStorage.clear();

    return (
        <div className={"fullscreen"}>
            <img src={"/ziggy.svg"} width={200}/>
            <h1>Welcome to zrok!</h1>

            <p>You can proceed to the <a href={"/"}>zrok web portal</a> and log in with your email and password.</p>

            <p>To enable your shell for zrok, use this command:</p>

            <pre>
                $ <span id={"zrok-enable-command"}>zrok enable {props.token}</span> <Icon path={mdiContentCopy} size={0.7} onClick={handleCopy}/>
            </pre>
        </div>
    )
}

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
    }, []);

    if(activeRequest) {
        step = <RegistrationForm
            email={email}
            token={token}
        />
    } else {
        step = <NoAccountRequest />
    }

    return (
        <div>{step}</div>
    )
}

export default Register;