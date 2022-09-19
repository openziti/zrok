import Login from './Login';
import Logout from './Logout';
import Network from './Network';
import Token from './Token';
import Version from './Version';
import { useEffect, useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./Register";

const Overview = () => {
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

    const logout = () => {
        setUser(null);
        localStorage.clear();
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
                            <Token user={user}/>
                            <Logout user={user} logout={logout}/>
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

const App = () => {
    return (
        <Router>
            <Routes>
                <Route path={"/"} element={<Overview />}/>
                <Route path={"register/:token"} element={<Register />} />
            </Routes>
        </Router>
    );
}

export default App;

