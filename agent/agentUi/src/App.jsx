import './App.css'
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import DataTable from 'react-data-table-component';

function App() {
    const [version, setVersion] = useState("");
    const [shares, setShares] = useState([]);

    const shareColumns = [
        {
            name: 'Token',
            selector: row => row.token
        }
    ];

    let api = new AgentApi(new ApiClient("http://localhost:8888"));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
           if(mounted) {
               setVersion(data.v);
           }
        });
    }, [api]);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                if(mounted) {
                    setShares(data.shares);
                }
            });
        }, 1000);
        return () => {
            mounted = false;
            clearInterval(interval);
        }
    });

    return (
        <>
            <h1>zrok Agent</h1>
            <code>{version}</code>

            <div>
                <h2>Shares</h2>
                <DataTable
                    columns={shareColumns}
                    data={shares}
                />
            </div>
        </>
    )
}

export default App;
