import './App.css'
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";
import DataTable from 'react-data-table-component';

function App() {
    const [version, setVersion] = useState("");
    const [shares, setShares] = useState([]);
    const [accesses, setAccesses] = useState([]);

    const shareColumns = [
        {
            name: 'Token',
            selector: row => row.reserved ? row.token+' (reserved)' : row.token
        },
        {
            name: 'Share Mode',
            selector: row => row.shareMode
        },
        {
            name: 'Backend Mode',
            selector: row => row.backendMode
        },
        {
            name: 'Frontend Endpoints',
            selector: row => row.frontendEndpoint
        },
        {
            name: 'Target',
            selector: row => row.backendEndpoint,
        },
        {
            name: 'Closed Permissions',
            selector: row => ''+row.closed
        }
    ];

    const accessColumns = [
        {
            name: 'Frontend Token',
            selector: row => row.frontendToken
        },
        {
            name: 'Token',
            selector: row => row.token
        },
        {
            name: 'Bind Address',
            selector: row => row.bindAddress
        },
    ];

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
           if(mounted) {
               setVersion(data.v);
           }
        });
    }, []);

    useEffect(() => {
        let mounted = true;
        let interval = setInterval(() => {
            api.agentStatus((err, data) => {
                if(mounted) {
                    setShares(data.shares);
                    setAccesses(data.accesses);
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

            <div>
                <h2>Accesses</h2>
                <DataTable
                    columns={accessColumns}
                    data={accesses}
                />
            </div>
        </>
    )
}

export default App;
