import Login from './Login';
import Logout from './Logout';
import Network from './Network';
import Version from './Version';
import {useEffect, useState} from "react";
import Environments from "./Environments";
import Icon from '@mdi/react';
import { mdiCloud } from '@mdi/js';

const App = () => {
    const [user, setUser] = useState();

    useEffect(() => {
        const localUser = localStorage.getItem("user")
        if (localUser) {
            setUser(JSON.parse(localUser))
            console.log('reloaded user', localUser)
        }
    }, []);

    if (!user) {
        return (
            <Login loginSuccess={setUser}/>
        );
    }

    return (
        <div className="zrok">
            <div className="container">
                <div className="header">
                    <img alt="ziggy goes to space" src="ziggy.svg" width="65px"/>
                    <p className="header-title">zrok</p>
                    <div className={"header-status"}>
                        <div>
                            <p>{user.email}</p>
                            <Version/>
                        </div>
                        <div className={"header-controls"}>
                            <button className={"logoutButton"}><Icon path={mdiCloud} size={0.7}/></button>
                            <Logout user={user} logout={() => {
                                setUser(null);
                                localStorage.clear();
                            }}/>
                        </div>
                    </div>
                </div>
                <div className="main">
                    <Network />
                </div>
            </div>
        </div>
    );
}

export default App;

