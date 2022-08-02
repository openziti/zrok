import Login from './Login';
import Logout from './Logout';
import Version from './Version';
import {useEffect, useState} from "react";
import Identities from "./Identities";

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
            <Login
                loginSuccess={setUser}
            />
        );
    }

    return (
        <div className="zrok">
            <div className="container">
                <div className="header">
                    <img src="ziggy.png" width="100px"/>
                    <p className="title">zrok</p>
                    <div class="header-left">
                        <div>
                            <Logout user={user} logout={() => {
                                setUser(null);
                                localStorage.clear();
                            }}/>
                        </div>
                        <div>
                            <Version/>
                        </div>
                    </div>
                </div>
                <div className="main">

                    <Identities user={user}/>
                </div>
            </div>
        </div>
    );
}

export default App;

