import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";

const AgentUi = (props) => {
    const [version, setVersion] = useState("");
    const [shares, setShares] = useState([]);
    const [accesses, setAccesses] = useState([]);

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

    const router = createBrowserRouter([
        {
            path: "/",
            element: <Overview version={version} shares={shares} accesses={accesses} />
        },
        {
            path: "/share/:token",
            element: <ShareDetail version={version} shares={shares} />
        }
    ]);

    return <RouterProvider router={router} />
}

export default AgentUi;