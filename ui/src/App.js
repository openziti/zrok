import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./Register";
import NewConsole from "./NewConsole";
import {useEffect, useState} from "react";
import Login from "./Login";

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

    const consoleComponent = user ? <NewConsole logout={logout} user={user} /> : <Login loginSuccess={setUser} />

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