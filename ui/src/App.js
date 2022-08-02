import Login from './Login';
import Logout from './Logout';
import Version from './Version';
import * as gateway from "./api/gateway";
import {useEffect, useState} from "react";

gateway.init({
   url: '/api/v1'
});

const App = () => {
    const [user, setUser] = useState();

    useEffect(() => {
        const localUser = localStorage.getItem("user")
        if(localUser) {
            setUser(JSON.parse(localUser))
            console.log('reloaded user', localUser)
        }
    }, []);

    if (!user) {
        return (
            <Login
                loginSuccess={setUser}
            />
        );
    }

    return (
        <div className="zrok">
            <header className="zrok-header">
                <h1>zrok</h1>
                <Version/>
                <Logout user={user} logout={() => { setUser(null); }}/>
            </header>
        </div>
    );
}

export default App;
