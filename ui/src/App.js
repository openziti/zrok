import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./register/Register";
import Console from "./console/Console";
import {useEffect, useState} from "react";
import Login from "./console/login/Login";
import ResetPassword from "./resetPassword/ResetPassword"

const App = () => {
    const [user, setUser] = useState();

    useEffect(() => {
        function checkUserData() {
            const localUser = localStorage.getItem("user");
            if(localUser) {
                console.log(localUser)
                setUser(JSON.parse(localUser));
                console.log("reloaded user", localUser);
            }
        }
      
        document.addEventListener('storage', checkUserData)
      
        return () => {
          document.removeEventListener('storage', checkUserData)
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
                <Route path={"resetPassword"} element={<ResetPassword />}/>
                <Route path={"resetPassword/:token"} element={<ResetPassword />}/>
            </Routes>
        </Router>
    );
}

export default App;