import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./Register";
import Overview from "./Overview";

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

