import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from "./Register";
import Console from "./Console";

const App = () => {
    return (
        <Router>
            <Routes>
                <Route path={"/"} element={<Console/>}/>
                <Route path={"register/:token"} element={<Register/>} />
            </Routes>
        </Router>
    );
}

export default App