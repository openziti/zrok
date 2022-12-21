import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./register/Register";
import Console from "./console/Console";
import {useEffect, useState} from "react";
import Login from "./login/Login";

const App = () => {
    const [user, setUser] = useState();

    useEffect(() => {
        const localUser = localStorage.getItem("user");
        if(localUser) {
            setUser(JSON.parse(localUser));
            console.log("reloaded user", localUser);
        }
    }, []);

    const logout = () => {
        setUser(null);
        localStorage.clear();
    }

    const consoleComponent = user ? <Console logout={logout} user={user} /> : <Login loginSuccess={setUser} />

    return (
        <Router>
            <Routes>
                <Route path={"/"} element={consoleComponent}/>
                <Route path={"register/:token"} element={<Register />} />
            </Routes>
        </Router>
    );
}

export default App;