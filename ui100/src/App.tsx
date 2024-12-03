import {BrowserRouter, Route, Routes} from "react-router";
import ApiConsole from "./ApiConsole.tsx";
import Login from "./Login.tsx";
import {useEffect, useState} from "react";
import {User} from "./model/user.ts";

const App = () => {
    const [user, setUser] = useState(null as User);

    useEffect(() => {
        const checkUser = () => {
            const user = localStorage.getItem("user");
            if (user) {
                console.log(user);
                setUser(JSON.parse(user));
                console.log("reloaded user", user);
            }
        }
        checkUser();
    }, []);

    const login = (user: User) => {
        setUser(user);
        localStorage.setItem("user", JSON.stringify(user));
    }

    const logout = () => {
        setUser(null);
        localStorage.clear();
    }

    const consoleRoot = user ? <ApiConsole user={user} logout={logout}/> : <Login onLogin={login}/>

    return (
        <BrowserRouter>
            <Routes>
                <Route index element={consoleRoot}/>
            </Routes>
        </BrowserRouter>
    );
}

export default App;