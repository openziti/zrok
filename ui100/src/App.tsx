import {BrowserRouter, Route, Routes} from "react-router";
import ApiConsole from "./ApiConsole.tsx";
import Login from "./Login.tsx";
import {useEffect} from "react";
import {User} from "./model/user.ts";
import useApiConsoleStore from "./model/store.ts";

const App = () => {
    const user = useApiConsoleStore((state) => state.user);
    const updateUser = useApiConsoleStore((state) => state.updateUser);

    useEffect(() => {
        const checkUser = () => {
            const user = localStorage.getItem("user");
            if (user) {
                updateUser(JSON.parse(user));
            }
        }
        checkUser();
    }, []);

    const login = (user: User) => {
        updateUser(user);
        localStorage.setItem("user", JSON.stringify(user));
    }

    const logout = () => {
        updateUser(null as User);
        localStorage.clear();
    }

    const consoleRoot = user ? <ApiConsole logout={logout}/> : <Login onLogin={login}/>

    return (
        <BrowserRouter>
            <Routes>
                <Route index element={consoleRoot}/>
            </Routes>
        </BrowserRouter>
    );
}

export default App;