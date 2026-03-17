import {BrowserRouter, Route, Routes} from "react-router";
import ApiConsole from "./ApiConsole.tsx";
import Login from "./Login.tsx";
import {useEffect} from "react";
import {clearStoredUser, loadStoredUser, saveStoredUser, User} from "./model/user.ts";
import useApiConsoleStore from "./model/store.ts";
import ForgotPassword from "./ForgotPassword.tsx";
import Register from "./Register.tsx";
import ResetPassword from "./ResetPassword.tsx";
import ErrorBoundary from "./ErrorBoundary.tsx";

const App = () => {
    const user = useApiConsoleStore((state) => state.user);
    const updateUser = useApiConsoleStore((state) => state.updateUser);

    useEffect(() => {
        const checkUser = () => {
            updateUser(loadStoredUser());
        };
        checkUser();

        document.addEventListener("userUpdated", checkUser);

        return () => {
            document.removeEventListener("userUpdated", checkUser);
        };
    }, [updateUser]);

    const login = (user: User) => {
        updateUser(user);
        saveStoredUser(user);
    };

    const logout = () => {
        updateUser(null);
        clearStoredUser();
    };

    const consoleRoot = user
        ? <ErrorBoundary><ApiConsole logout={logout}/></ErrorBoundary>
        : <Login onLogin={login}/>

    return (
        <BrowserRouter>
            <Routes>
                <Route index element={consoleRoot}/>
                <Route path="/forgotPassword" element={<ErrorBoundary><ForgotPassword /></ErrorBoundary>} />
                <Route path="/register/:regToken" element={<ErrorBoundary><Register /></ErrorBoundary>} />
                <Route path="/resetPassword/:resetToken" element={<ErrorBoundary><ResetPassword /></ErrorBoundary>} />
            </Routes>
        </BrowserRouter>
    );
};

export default App;
