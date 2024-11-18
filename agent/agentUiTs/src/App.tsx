import './App.css'
import {useEffect, useState} from 'react'
import {AgentApi, Configuration} from "./api";

function App() {
    const [version, setVersion] = useState("not set");
    const [errorMessage, setErrorMessage] = useState("no error");

    useEffect(() => {
        let api = new AgentApi(new Configuration({basePath: window.location.origin}));
        api.agentVersion().then((v) => {
            if(v.v) {
                setVersion(v.v);
            } else {
                console.log(v);
            }
        }).catch((v) => {
           console.log("caught", v.toString());
        });
    }, []);

    useEffect(() => {
        let api = new AgentApi(new Configuration({basePath: window.location.origin}));
        api.agentAccessPrivate({token: "a", bindAddress: "127.0.0.1:9911"}).catch(e => {
            console.log(e.response.json().then(eb => {
                setErrorMessage(eb.message);
                console.log(eb.message);
            }));
        }).then(v => {
            console.log(v);
        });
    }, []);

    return (
        <>
            <h1>Agent UI</h1>
            <h2>{version}</h2>
            <h4>{errorMessage}</h4>
        </>
    )
}

export default App
