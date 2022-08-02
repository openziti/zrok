import Version from './Version';
import * as gateway from "./api/gateway";

gateway.init({
   url: '/api/v1'
});

const App = () => {
    return (
        <div className="zrok">
            <header className="zrok-header">
                <h1>zrok</h1>
                <Version/>
            </header>
        </div>
    );
}

export default App;
